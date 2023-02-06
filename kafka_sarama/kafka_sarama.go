package kafkasarama

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/suiguo/hwlib/logger"
)

const KafkaSaramaTag = "KafkaSarama"

type Producer interface {
	PushMsg(topic string, msg string) error
	Close() error
}

// 同步生产者
type syncproducer struct {
	sarama.SyncProducer
	log logger.Logger
}

func (a *syncproducer) PushMsg(topic string, msg string) error {
	if a.SyncProducer == nil {
		return fmt.Errorf("SyncProducer is nil")
	}
	productMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(msg)}
	_, _, err := a.SendMessage(productMsg)
	if err != nil {
		if a.log != nil {
			a.log.Error(KafkaSaramaTag, "PushMsg", err)
		}
		return err
	}
	return nil
}
func (a *syncproducer) Close() error {
	if a.SyncProducer != nil {
		return a.SyncProducer.Close()
	}
	return nil
}

type asyncproducer struct {
	sarama.AsyncProducer
	log logger.Logger
}

func (a *asyncproducer) PushMsg(topic string, msg string) error {
	if a.AsyncProducer == nil {
		return fmt.Errorf("SyncProducer is nil")
	}
	productMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(msg)}
	select {
	case a.AsyncProducer.Input() <- productMsg:
		return nil
	case err := <-a.AsyncProducer.Errors():
		return err
	}
}
func (a *asyncproducer) Close() error {
	if a.AsyncProducer != nil {
		return a.AsyncProducer.Close()
	}
	return nil
}

type ProductConfig func(*sarama.Config)

// ack
func WithProductAcks(ack sarama.RequiredAcks) ProductConfig {
	return func(c *sarama.Config) {
		c.Producer.RequiredAcks = ack
	}
}

// 超时
func WithProductTimeOut(t time.Duration) ProductConfig {
	return func(c *sarama.Config) {
		c.Producer.Timeout = t
	}
}

func WithProductReTryTimes(max int) ProductConfig {
	return func(c *sarama.Config) {
		c.Producer.Retry.Max = max
	}
}

// 地址  是否是同步 配置
func NewSarProducer(addrs []string, is_sync bool, log logger.Logger, cfg ...ProductConfig) (Producer, error) {
	config := sarama.NewConfig()
	for _, c := range cfg {
		c(config)
	}
	if is_sync {
		config.Producer.Return.Successes = true
		p, err := sarama.NewSyncProducer(addrs, config)
		if err == nil {
			return &syncproducer{SyncProducer: p,
				log: log}, nil
		}
		return nil, err
	}
	p, err := sarama.NewAsyncProducer(addrs, config)
	if err == nil {
		return &asyncproducer{AsyncProducer: p,
			log: log}, nil
	}
	// sarama.ConsumerMessage
	return nil, err
}

//消费者

type Consumer interface {
	SubscribeTopics([]string, Handler)
	Close() error
}
type comsumer struct {
	sarama.ConsumerGroup
	log logger.Logger
	context.Context
	autoCommit bool
	topics     sync.Map //map[string]Handler
}
type Handler func(topic string, partition int32, offset int64, msg []byte) error

func (c *comsumer) SubscribeTopics(topic []string, h Handler) {
	for _, val := range topic {
		_, ok := c.topics.Load(val)
		if !ok {
			go func(t string) {
				for {
					err := c.Consume(c.Context, []string{t}, c)
					if c.log != nil {
						c.log.Error(KafkaSaramaTag, "err", err)
					}
				}
			}(val)
		}
		c.topics.Store(val, h)
	}
}
func (c *comsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (c *comsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (c *comsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.topics.Range(func(key, value any) bool {
			if msg.Topic == key {
				h := value.(Handler)
				if h(msg.Topic, msg.Partition, msg.Offset, msg.Value) == nil {
					if c.autoCommit {
						sess.MarkMessage(msg, "")
					} else {
						sess.Commit()
					}
				}
			}
			return true
		})
	}
	return nil
}

// func (c *comsumer) Close() error {
// 	return c.Close()
// }

type ConsumerConfig func(*sarama.Config)

func WithConsumerAutoCommit(is_auto bool) ConsumerConfig {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.AutoCommit.Enable = is_auto
	}
}

func WithConsumerAutoInterval(Interval time.Duration) ConsumerConfig {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.AutoCommit.Interval = Interval
	}
}

type OffsetType int64

const (
	OffsetOldest OffsetType = OffsetType(sarama.OffsetOldest)
	OffsetNewest OffsetType = OffsetType(sarama.OffsetNewest)
)

func WithConsumerOffsets(i OffsetType) ConsumerConfig {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.Initial = int64(i)
	}
}

func NewSarConsumer(addrs []string, group string, log logger.Logger, cfg ...ConsumerConfig) (Consumer, error) {
	config := sarama.NewConfig()
	// config.Consumer.Offsets.AutoCommit.Enable = false
	for _, c := range cfg {
		c(config)
	}
	config.Version = sarama.DefaultVersion //  V1_0_0_0
	config.Consumer.Return.Errors = true
	client, err := sarama.NewClient(addrs, config)
	if err != nil {
		return nil, err
	}
	group_client, err := sarama.NewConsumerGroupFromClient(group, client)
	if err != nil {
		return nil, err
	}
	go func() {
		for err := range group_client.Errors() {
			if log != nil {
				log.Error(KafkaSaramaTag, "err", err)
			}
		}
	}()
	return &comsumer{ConsumerGroup: group_client,
		log:        log,
		Context:    context.Background(),
		autoCommit: config.Consumer.Offsets.AutoCommit.Enable,
	}, nil
}
