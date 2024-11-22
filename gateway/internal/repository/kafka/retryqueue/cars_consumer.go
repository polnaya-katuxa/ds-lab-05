package retryqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/clients"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CarsRetryQueueConsumer struct {
	consumer sarama.ConsumerGroup
	logger   *zap.SugaredLogger

	topic string
}

func NewCarsRetryQueueConsumer(
	ctx context.Context,
	producer *RetryQueueProducer,
	brokers []string,
	topic string,
	cars *clients.CarsServiceClient,
	logger *zap.SugaredLogger,
) (*CarsRetryQueueConsumer, error) {
	sl, _ := zap.NewStdLogAt(logger.Desugar(), zapcore.WarnLevel)
	sarama.Logger = sl

	config := sarama.NewConfig()
	config.ClientID = "car-rental-system"

	consumerGroup, err := sarama.NewConsumerGroup(brokers, config.ClientID, config)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	consumer := &CarsRetryQueueConsumer{
		consumer: consumerGroup,
		logger:   logger,
		topic:    topic,
	}

	carUnbookConsumer := &carUnbookConsumer{
		ready:    make(chan bool),
		cars:     cars,
		producer: producer,
		logger:   logger,
	}

	consumer.retryCarUnbook(ctx, carUnbookConsumer)

	return consumer, nil
}

func (q *CarsRetryQueueConsumer) Stop() {
	q.consumer.Close()
}

type carUnbookConsumer struct {
	ready    chan bool
	producer *RetryQueueProducer
	cars     *clients.CarsServiceClient
	logger   *zap.SugaredLogger
}

func (q *CarsRetryQueueConsumer) retryCarUnbook(ctx context.Context, consumer *carUnbookConsumer) {
	go func() {
		for {
			if err := q.consumer.Consume(ctx, []string{q.topic}, consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				continue
			}

			if ctx.Err() != nil {
				return
			}

			consumer.ready = make(chan bool)
		}
	}()

	q.logger.Info("waiting for car retries consumer")

	<-consumer.ready

	q.logger.Info("car retries consumer ready")
}

func (c *carUnbookConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *carUnbookConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *carUnbookConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				c.logger.Warnw("message channel was closed")
				return nil
			}

			c.logger.Infow("message claimed", "value", string(message.Value), "timestamp", message.Timestamp, "topic", message.Topic)

			var carUnbookRetryMsg models.CarUnbookRetryMsg
			err := json.Unmarshal(message.Value, &carUnbookRetryMsg)
			if err != nil {
				session.MarkMessage(message, "msg has invalid body")
				c.logger.Errorw("unmarshal CarUnbookRetryMsg", "error", err)
				continue
			}

			for time.Now().Sub(message.Timestamp) < time.Second*10 {
			}

			err = c.cars.RetryUnbook(session.Context(), carUnbookRetryMsg.CarUid)
			if err != nil {
				c.logger.Warnw("cannot cancel car book", "car", carUnbookRetryMsg.CarUid, "error", err)
				session.MarkMessage(message, "retry cancel car book")
				c.producer.RetryCarUnbook(carUnbookRetryMsg.CarUid)
				continue
			}

			c.logger.Infow("cancelled car book", "car", carUnbookRetryMsg.CarUid)
			session.MarkMessage(message, "got msg for car unbook")

		case <-session.Context().Done():
			return nil
		}
	}
}
