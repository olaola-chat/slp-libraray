package kafka

import (
	"time"

	"github.com/gogf/gf/frame/g"

	"github.com/Shopify/sarama"
	"github.com/gogf/gf/util/gconv"
	"github.com/silenceper/pool"
)

// NewClientWithPool 根据 name 实例创建一个kafka连接池
func NewClientWithPool(name string) *Client {
	config, err := GetConfig(name)
	if err != nil {
		panic(err)
	}
	version, err := sarama.ParseKafkaVersion(config.Version)
	if err != nil {
		panic(err)
	}

	factory := func() (interface{}, error) {
		cfg := sarama.NewConfig()
		cfg.Producer.RequiredAcks = sarama.WaitForLocal
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
		cfg.Producer.Return.Successes = true
		cfg.Producer.Return.Errors = true
		cfg.Version = version
		return sarama.NewSyncProducer(config.Host, cfg)
	}
	close := func(v interface{}) error {
		return v.(sarama.SyncProducer).Close()
	}

	poolConfig := &pool.Config{
		InitialCap: 1,  //资源池初始连接数
		MaxIdle:    20, //最大空闲连接数
		MaxCap:     50, //最大并发连接数
		Factory:    factory,
		Close:      close,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: 60 * time.Second,
	}

	kafkaPool, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		panic(err)
	}
	client := &Client{
		Config: config,
		Pool:   kafkaPool,
	}

	return client
}

// Client kafka连接池对象
type Client struct {
	Config *Config
	Pool   pool.Pool
}

// Send 发送消息，更快捷方式见下面
func (serv *Client) Send(topic string, value sarama.Encoder, key ...interface{}) (int32, int64, error) {
	message := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     value,
		Timestamp: time.Now(),
	}

	if len(key) > 0 {
		hashKey := gconv.String(key[0])
		message.Key = sarama.StringEncoder(hashKey)
	}

	v, err := serv.Pool.Get()
	if err != nil {
		return 0, 0, err
	}
	prod, _ := v.(sarama.SyncProducer)

	partition, offset, err := prod.SendMessage(message)

	if err != nil {
		g.Log().Error("Kafka send message error", err)
		serv.Pool.Close(prod)
		return 0, 0, err
	} else {
		err := serv.Pool.Put(prod)
		if err != nil {
			g.Log().Error("Kafka Pool put error", err)
		}
	}

	return partition, offset, err
}

// SendString 发送字符串消息
func (serv *Client) SendString(topic string, value string, key ...interface{}) (int32, int64, error) {
	return serv.Send(topic, sarama.StringEncoder(value), key...)
}

// SendStringBytes 发送[]byte消息
func (serv *Client) SendStringBytes(topic string, value []byte, key ...interface{}) (int32, int64, error) {
	return serv.Send(topic, sarama.ByteEncoder(value), key...)
}
