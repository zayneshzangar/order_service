package service

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

const (
	sessionTimeout = 7000 // ms
	noTimeout      = -1
)

type Hundler interface {
	HandleMessage(message []byte, topic kafka.TopicPartition, cn int64) error
}

type Consumer struct {
	consumer       *kafka.Consumer
	handler        Hundler
	stop           bool
	consumerNumber int64
}

func NewConsumer(handler Hundler, address []string, topic, consumerGroup string, consumerNumber int64) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(address, ","),
		"group.id":                 consumerGroup,
		"session.timeout.ms":       sessionTimeout,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest",
	}
	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, err
	}
	return &Consumer{
		consumer:       c,
		handler:        handler,
		consumerNumber: consumerNumber,
	}, nil
}

func (c *Consumer) Start() {
	for {
		if c.stop {
			break
		}
		
		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			logrus.Error(err)
			continue
		}

		if err := c.handler.HandleMessage(kafkaMsg.Value, kafkaMsg.TopicPartition, c.consumerNumber); err != nil {
			logrus.Error(err)
			continue
		}
		
		if _, err := c.consumer.StoreMessage(kafkaMsg); err != nil {
			logrus.Error(err)
			continue
		}
		
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}

	logrus.Infof("Commited offset")
	return c.consumer.Close()
}
