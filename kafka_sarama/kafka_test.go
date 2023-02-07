package kafkasarama

import (
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/suiguo/hwlib/logger"
)

func TestXxx(t *testing.T) {
	p, err := NewSarProducer([]string{"localhost:9092"}, true, logger.NewStdLogger(), WithProductAcks(sarama.WaitForAll))
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < 100; i++ {
		err = p.PushMsg("topic", []byte("hello"))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func TestAsync(t *testing.T) {
	p, err := NewSarConsumer([]string{"localhost:9092"}, "testgroup", logger.NewStdLogger(),
		WithConsumerAutoCommit(true),
		WithConsumerOffsets(OffsetOldest))
	if err != nil {
		fmt.Println(err)
	}
	p.SubscribeTopics([]string{"topic0", "topic1", "topic2"}, func(topic string, partition int32, offset int64, msg []byte) error {
		fmt.Println(topic, partition, offset, string(msg))
		return nil
	})
	// fmt.Println(err)
	k := 0
	for {
		time.Sleep(time.Second * 3)
		p, err := NewSarProducer([]string{"localhost:9092"}, true, logger.NewStdLogger(), WithProductAcks(sarama.WaitForAll))
		if err != nil {
			fmt.Println(err)
		}
		for i := 0; i < 100; i++ {
			k++
			if k < 500 {
				err = p.PushMsg(fmt.Sprintf("topic%d", k%3), []byte(fmt.Sprintf("hello%d", k)))
				if err != nil {
					fmt.Println(err)
				}
			}
		}

	}
}
