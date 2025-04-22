package binlog

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"

	"github.com/olaola-chat/slp-library/kafka"
)

type ConsumerConf = kafka.ConsumerConf

func NewWatchWorkers(
	cfg *ConsumerConf, count int, tables map[string]Callback) ([]*kafka.Worker, error) {

	if count < 1 {
		count = 1
	}

	wrap := cbWrap{tables: tables}

	workers := make([]*kafka.Worker, 0, count)
	for i := 0; i < count; i++ {
		worker, err := kafka.NewConsumerWorker(cfg, wrap.receiveMsg)
		if err != nil {
			return nil, err
		}
		workers = append(workers, worker)
	}

	return workers, nil
}

func RunWorkers(workers []*kafka.Worker) {
	if len(workers) == 0 {
		return
	}

	count := len(workers)

	wg := &sync.WaitGroup{}
	stops := make([]chan bool, 0)

	for i := 0; i < count; i++ {
		stop := make(chan bool, 1)
		stops = append(stops, stop)
		_ = workers[i].Start(stop, wg)
	}

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	v := <-sign
	g.Log().Println("RunWorkers receive sign", v)
	for _, stop := range stops {
		stop <- true
	}

	wg.Wait()
	g.Log().Println("RunWorkers completely closed")
}

type cbWrap struct {
	tables map[string]Callback
}

func (c cbWrap) receiveMsg(msg *sarama.ConsumerMessage) error {
	now := time.Now().UnixNano()

	//ctx := myctx.NewContext(msg.Topic)
	ctx := context.Background()

	kafkaTime := msg.Timestamp.UnixNano()
	kafkaDur := float64(now-kafkaTime) / 1e6 //kafka到now的时间段

	res, err := ParseCanalJSON(msg.Value)
	if err != nil {
		g.Log().Errorf(
			"watchdb, topic:%s,partition:%d,offset:%d,"+
				"kafka dur(ms):%f,cost_time(ms):%.2f,err:%v",
			msg.Topic, msg.Partition, msg.Offset,
			kafkaDur, float64(time.Now().UnixNano()-now)/1e6, err)
		return nil
	}

	//dbDur1 := res.Ts - res.Es
	//dbDur2 := kafkaTime/1e6 - res.Ts

	count := len(res.Data)
	if count == 0 {
		count = len(res.Old)
	}

	/*
		dealed := false

		defer func() {
			if dealed {
				g.Log().Printf(
					"watchdb, topic:%s,partition:%d,offset:%d,"+
						"database:%s,table:%s,op:%s,"+
						"db dur(ms):%d/%d,kafka dur(ms):%.2f,cost_time(ms):%.2f,"+
						"count:%d,dealed:%v,success:%v",
					msg.Topic, msg.Partition, msg.Offset,
					res.Database, res.Table, res.Op,
					dbDur1, dbDur2, kafkaDur, float64(time.Now().UnixNano()-now)/1e6,
					count, dealed, err == nil)
			}
			//g.Log().Printf("old: %v, new: %v", res.Old, res.Data)
		}()
	*/

	cb, ok := c.tables[res.Table]
	if !ok {
		return nil
	}
	//dealed = true

	switch res.Op {
	case CanalWrite:
		values := make([]interface{}, 0, count)
		for _, v := range res.Data {
			values = append(values, buildProtoMessage(cb, v))
		}
		_ = cb.Inserted(ctx, values, res)
	case CanalUpdate:
		values := make([]interface{}, 0, count)
		for _, v := range res.Data {
			values = append(values, buildProtoMessage(cb, v))
		}
		_ = cb.Updated(ctx, values, res)
	case CanalDelete:
		values := make([]interface{}, 0, count)
		for _, v := range res.Old {
			values = append(values, buildProtoMessage(cb, v))
		}
		_ = cb.Deleted(ctx, values, res)
	}

	return nil
}

func buildProtoMessage(cb Callback, data map[string]string) interface{} {
	val := cb.New()
	err := gconv.Struct(data, val)
	if err != nil {
		g.Log().Printf("buildProtoMessage failed, %v, err:%v", data, err)
		return nil
	}
	return val
}
