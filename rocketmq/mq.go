package rocketmq

import (
	"context"
	"fmt"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/frame/gins"
)

type Client struct {
	Config *Config
}

func NewClient(name string) *Client {
	instanceKey := fmt.Sprintf("self-go-rocketmq.%s", name)
	result := gins.GetOrSetFuncLock(instanceKey, func() interface{} {
		config := Config{}
		err := g.Cfg().GetStruct(fmt.Sprintf("go-rocketmq.%s", name), &config)
		if err != nil {
			panic(gerror.Wrap(err, "rocketmq config error"))
		}
		return &Client{
			Config: &config,
		}
	})
	if client, ok := result.(*Client); ok {
		return client
	}
	//理论上是不可能到这一步的
	panic(gerror.New("get rocketmq client error"))
}

func (c *Client) Produce(topic string, body []byte) error {
	producer, err := rmq_client.NewProducer(&rmq_client.Config{
		Endpoint: c.Config.EndPoint,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    c.Config.AccessKey,
			AccessSecret: c.Config.AccessSecret,
		},
	})
	if err != nil {
		g.Log().Errorf("rocketmq produce error||topic=%s||body=%s", topic, string(body))
		return err
	}
	err = producer.Start()
	if err != nil {
		g.Log().Errorf("rocketmq produce start error||topic=%s||body=%s", topic, string(body))
		return err
	}
	defer producer.GracefulStop()
	msg := &rmq_client.Message{
		Topic: topic,
		Body:  body,
	}
	_, err = producer.Send(context.TODO(), msg)
	if err != nil {
		g.Log().Errorf("rocketmq produce send error||topic=%s||body=%s", topic, string(body))
		return err
	}
	return nil
}
