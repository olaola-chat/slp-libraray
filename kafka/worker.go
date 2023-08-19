package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gogf/gf/frame/g"
)

// ConsumerTopics 消费者参数定义
type ConsumerTopics struct {
	Name      string   //kafka 集群配置名字
	GroupName string   //消费组名字
	Topics    []string //对应topic数据
	Desc      string   //说明信息
}

// ConsumerConfig 消费者参数定义
type ConsumerConfig struct {
	Cfg    *ConsumerTopics
	Hander ConsumerHander //回调函数
	Stop   chan bool      //停止通知
	Wait   *sync.WaitGroup
}

// ConsumerHander 消息回调定义
type ConsumerHander func(msg *sarama.ConsumerMessage) error

// NewConsumerWorker kafka消费
func NewConsumerWorker(consumer *ConsumerConfig) {
	defer (func() {
		if consumer.Wait != nil {
			consumer.Wait.Done()
		}
	})()
	cfg, err := GetConfig(consumer.Cfg.Name)
	if err != nil {
		panic(err)
	}
	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(
		cfg.Host,
		consumer.Cfg.GroupName,
		config,
	)
	if err != nil {
		return
	}

	// Track errors
	// 这会不会成为一个死的....
	go func() {
		for err := range group.Errors() {
			g.Log().Info("ERROR", err)
		}
	}()

	consumerWithHander := Consumer{
		Hander: consumer.Hander,
	}

	closed := false
	go func() {
		for {
			if closed {
				return
			}
			err := group.Consume(context.Background(), consumer.Cfg.Topics, consumerWithHander)
			g.Log().Info("group.break", err)
			time.Sleep(time.Second)
		}
	}()
	g.Log().Info("kafka wait for control")

	<-consumer.Stop
	g.Log().Info("kafka to close")
	closed = true
	if err = group.Close(); err != nil {
		g.Log().Info("Error closing client: %v", err)
	}
	g.Log().Info("kafka closed")
}

// Consumer kafka消费定义
type Consumer struct {
	Hander ConsumerHander
}

// Setup kafka连接后回调
func (serv Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	g.Log().Info("kafka to Setup")
	return nil
}

// Cleanup kafka关闭后回调
func (serv Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	g.Log().Info("kafka to Cleanup")
	return nil
}

// ConsumeClaim kafka消费回调
func (serv Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := serv.Hander(msg)
		if err != nil {
			return err
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
