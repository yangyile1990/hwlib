package kafkasarama

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/suiguo/hwlib/logger"
)

const KafkaSaramaTag = "KafkaSarama"

type Producer interface {
	PushMsg(topic string, msg []byte) error
	Close() error
}

// 同步生产者
type syncproducer struct {
	sarama.SyncProducer
	log logger.Logger
}

func (a *syncproducer) PushMsg(topic string, msg []byte) error {
	if a.SyncProducer == nil {
		return fmt.Errorf("SyncProducer is nil")
	}
	productMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(msg)}
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

func (a *asyncproducer) PushMsg(topic string, msg []byte) error {
	if a.AsyncProducer == nil {
		return fmt.Errorf("SyncProducer is nil")
	}
	productMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(msg)}
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
	SubscribeTopics([]string, Handler) error
	Close(topic string) error
}
type comsumer struct {
	addr  []string
	group string
	cfg   *sarama.Config
	log   logger.Logger
	context.Context
	autoCommit bool
	topics     sync.Map //map[string]Handler
}
type Handler func(topic string, partition int32, offset int64, msg []byte) error
type topicHandler struct {
	h Handler
	sarama.ConsumerGroup
}

func (c *comsumer) SubscribeTopics(topic []string, h Handler) error {
	str := ""
	for _, val := range topic {
		str += fmt.Sprintf("[%s]", val)
	}
	if str == "" {
		return fmt.Errorf("no topic")
	}
	_, ok := c.topics.Load(str)
	if !ok {
		client, err := sarama.NewClient(c.addr, c.cfg)
		if err != nil {
			return err
		}
		group_client, err := sarama.NewConsumerGroupFromClient(c.group, client)
		if err != nil {
			return err
		}
		go func() {
			for err := range group_client.Errors() {
				if c.log != nil {
					c.log.Error(KafkaSaramaTag, "err", err)
				}
			}
		}()
		go func(t []string, cli sarama.ConsumerGroup) {
			for {
				err := cli.Consume(c.Context, t, c)
				if c.log != nil {
					c.log.Error(KafkaSaramaTag, "err", err)
				}
			}
		}(topic, group_client)
		c.topics.Store(str, &topicHandler{h: h, ConsumerGroup: group_client})
	}
	return nil
}

func (c *comsumer) Close(topic string) error {
	cli, ok := c.topics.Load(topic)
	if ok {
		t, ok := cli.(sarama.ConsumerGroup)
		if ok {
			return t.Close()
		}
		return fmt.Errorf("cant conver to ConsumerGroup")
	}
	return nil
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
			k := key.(string)
			if strings.Contains(k, fmt.Sprintf("[%s]", msg.Topic)) {
				h := value.(*topicHandler)
				if h.h(msg.Topic, msg.Partition, msg.Offset, msg.Value) == nil {
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
	return &comsumer{
		group:      group,
		addr:       addrs,
		cfg:        config,
		log:        log,
		Context:    context.Background(),
		autoCommit: config.Consumer.Offsets.AutoCommit.Enable,
	}, nil
}
