package rocketmq

import (
	"context"
	"testing"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
)

func TestRocket(t *testing.T) {
	producer, err := rmq_client.NewProducer(&rmq_client.Config{
		Endpoint:    "127.0.0.1:8082",
		Credentials: &credentials.SessionCredentials{},
	})
	if err != nil {
		panic(err)
	}
	err = producer.Start()
	if err != nil {
		panic(err)
	}
	defer producer.GracefulStop()
	msg := &rmq_client.Message{
		Topic: "test",
		Body:  []byte("this is a message : hbc"),
	}
	ctx, _ := context.WithTimeout(context.TODO(), 5*time.Second)
	_, err = producer.Send(ctx, msg)
	if err != nil {
		panic(err)
	}
}
