package kafka

import (
	"fmt"
	"sync"
	"time"

	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/suiguo/hwlib/logger"
)

type KafkaType string

const KafkaLogTag = "Kafka"
const (
	ALLType      KafkaType = "all"
	ProducerType KafkaType = "Produce"
	ConsumerType KafkaType = "Consumer"
)

type KafkaMsg struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       string
	Msg       []byte
	MetaData  string
}
type KafaClient interface {
	Producer
	Consumer
}
type Producer interface {
	Produce(topic string, msg *KafkaMsg) error
}
type Consumer interface {
	MessageChan() <-chan *kafka.Message
	Subscribe(...string) error
}

type kafkaClient struct {
	sync.Once
	msgPopChan chan *kafka.Message
	logger.Logger
	producer *kafka.Producer
	consumer *kafka.Consumer
}

func (k *kafkaClient) Subscribe(topics ...string) error {
	if k.consumer == nil {
		return fmt.Errorf("consumer not int")
	}
	return k.consumer.SubscribeTopics(topics, nil)
}
func (k *kafkaClient) MessageChan() <-chan *kafka.Message {
	return k.msgPopChan
}
func (k *kafkaClient) Produce(topic string, msg *KafkaMsg) error {
	if k.producer == nil {
		return fmt.Errorf("producer not init")
	}
	topic_partition := kafka.TopicPartition{}
	if msg.MetaData != "" {
		topic_partition.Metadata = &msg.MetaData
	}
	if msg.Topic != "" {
		topic_partition.Topic = &msg.Topic
	}
	topic_partition.Offset.Set(msg.Offset)
	return k.producer.Produce(&kafka.Message{
		TopicPartition: topic_partition,
		Key:            []byte(msg.Key),
		Value:          msg.Msg,
	}, nil)
}

// run
func (k *kafkaClient) run() {
	k.Once.Do(func() {
		for {
			if k.consumer == nil {
				time.Sleep(time.Second * 2)
			}
			msg, err := k.consumer.ReadMessage(time.Second)
			if err != nil {
				if k.Logger != nil {
					k.Logger.Error(KafkaLogTag, "ReadMessage", err)
				} else {
					log.Println(KafkaLogTag, "ReadMessage", err)
				}
				time.Sleep(time.Second * 2)
			}
			go func() {
				k.msgPopChan <- msg
			}()
		}
	})
}
func GetKafkaByCfg(ktype KafkaType, consumer kafka.ConfigMap, producer kafka.ConfigMap, log logger.Logger) (KafaClient, error) {
	tmp := &kafkaClient{
		Logger:     log,
		msgPopChan: make(chan *kafka.Message, 1000),
	}
	var err error
	switch ktype {
	case ALLType:
		tmp.consumer, err = kafka.NewConsumer(&consumer)
		if err != nil {
			return nil, err
		}
		tmp.producer, err = kafka.NewProducer(&producer)
	case ConsumerType:
		tmp.consumer, err = kafka.NewConsumer(&consumer)
	case ProducerType:
		tmp.producer, err = kafka.NewProducer(&producer)
	}
	if err != nil {
		return nil, err
	}
	tmp.run()
	return tmp, err
}
func GetDefaultKafka(ktype KafkaType, server string, group_id string, offset string, log logger.Logger) (KafaClient, error) {
	tmp := &kafkaClient{
		Logger:     log,
		msgPopChan: make(chan *kafka.Message, 1000),
	}
	var err error
	switch ktype {
	case ALLType:
		tmp.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": server,
			"group.id":          group_id,
			"auto.offset.reset": offset,
		})
		if err != nil {
			return nil, err
		}
		tmp.producer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": server,
		})
	case ConsumerType:
		tmp.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": server,
			"group.id":          group_id,
			"auto.offset.reset": offset,
		})
	case ProducerType:
		tmp.producer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": server,
		})
	}
	if err != nil {
		return nil, err
	}
	tmp.run()
	return tmp, err
}
