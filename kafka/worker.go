package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gogf/gf/frame/g"
)

type ConsumerConf struct {
	Name      string   //kafka 集群配置名字
	GroupName string   //消费组名字
	Topics    []string //对应topic数据
	Desc      string   //说明信息
}

// ConsumerConfig 消费者参数定义
type Worker struct {
	cfg       *ConsumerConf
	receiveCb ReceiveFunc //回调函数
	group     sarama.ConsumerGroup
}

// ConsumerHander 消息回调定义
type ReceiveFunc func(msg *sarama.ConsumerMessage) error

// NewConsumerWorker kafka消费
func NewConsumerWorker(cfg *ConsumerConf, handler ReceiveFunc) (*Worker, error) {
	conf, err := GetConfig(cfg.Name)
	if err != nil {
		g.Log().Printf("read kafka config failed, %v", err)
		return nil, err
	}
	version, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		g.Log().Printf(" sarama.ParseKafkaVersion failed,version:%s, err:%v", conf.Version, err)
		return nil, err
	}

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Version = version
	kafkaCfg.Consumer.Return.Errors = true
	kafkaCfg.Consumer.Offsets.CommitInterval = 1 * time.Second
	kafkaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(conf.Host, cfg.GroupName, kafkaCfg)
	if err != nil {
		g.Log().Printf("sarama.NewConsumerGroup failed, %v", err)
		return nil, err
	}

	return &Worker{cfg: cfg, receiveCb: handler, group: group}, nil
}

func (w *Worker) Start(stop <-chan bool, wait *sync.WaitGroup) error {
	if wait != nil {
		wait.Add(1)
	}

	consumer := Consumer{
		Handler: w.receiveCb,
	}

	closed := false
	go func() {
		for {
			if closed {
				return
			}
			err := w.group.Consume(context.Background(), w.cfg.Topics, consumer)
			if err != nil {
				g.Log().Printf("group.break, %s, %v", w.cfg.GroupName, err)
			}
			time.Sleep(time.Second)
		}
	}()

	go func() {
		defer func() {
			if wait != nil {
				wait.Done()
			}
		}()

		g.Log().Printf("worker wait for control, %s", w.cfg.GroupName)
		for {
			select {
			case <-stop:
				g.Log().Printf("close worker, %s", w.cfg.GroupName)
				closed = true
				if err := w.group.Close(); err != nil {
					g.Log().Printf("Error closing client: %v", err)
				}
				g.Log().Printf("worker closed, %s", w.cfg.GroupName)
				return
			case err := <-w.group.Errors(): // Track errors
				g.Log().Printf("[ERROR]%s, err:%v", w.cfg.GroupName, err)
			}
		}
	}()

	return nil
}

// Consumer kafka消费定义
type Consumer struct {
	Handler ReceiveFunc
}

// Setup kafka连接后回调
func (serv Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	//g.Log().Println("kafka to Setup")
	return nil
}

// Cleanup kafka关闭后回调
func (serv Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	//g.Log().Println("kafka to Cleanup")
	return nil
}

// ConsumeClaim kafka消费回调
func (serv Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := serv.Handler(msg)
		if err != nil {
			return err
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
