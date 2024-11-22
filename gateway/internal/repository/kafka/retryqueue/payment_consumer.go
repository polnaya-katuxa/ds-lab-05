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

type PaymentRetryQueueConsumer struct {
	consumer sarama.ConsumerGroup
	logger   *zap.SugaredLogger

	topic string
}

func NewPaymentRetryQueueConsumer(
	ctx context.Context,
	producer *RetryQueueProducer,
	brokers []string,
	topic string,
	payment *clients.PaymentServiceClient,
	logger *zap.SugaredLogger,
) (*PaymentRetryQueueConsumer, error) {
	sl, _ := zap.NewStdLogAt(logger.Desugar(), zapcore.WarnLevel)
	sarama.Logger = sl

	config := sarama.NewConfig()
	config.ClientID = "car-rental-system"

	consumerGroup, err := sarama.NewConsumerGroup(brokers, config.ClientID, config)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	consumer := &PaymentRetryQueueConsumer{
		consumer: consumerGroup,
		logger:   logger,
		topic:    topic,
	}

	paymentCancelConsumer := &paymentCancelConsumer{
		ready:    make(chan bool),
		payment:  payment,
		producer: producer,
		logger:   logger,
	}

	consumer.retryPaymentCancel(ctx, paymentCancelConsumer)

	return consumer, nil
}

func (q *PaymentRetryQueueConsumer) Stop() {
	q.consumer.Close()
}

type paymentCancelConsumer struct {
	ready    chan bool
	producer *RetryQueueProducer
	payment  *clients.PaymentServiceClient
	logger   *zap.SugaredLogger
}

func (q *PaymentRetryQueueConsumer) retryPaymentCancel(ctx context.Context, consumer *paymentCancelConsumer) {
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

	q.logger.Info("waiting for payment retries consumer")

	<-consumer.ready

	q.logger.Info("payment retries consumer ready")
}

func (c *paymentCancelConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *paymentCancelConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *paymentCancelConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				c.logger.Warnw("message channel was closed")
				return nil
			}

			c.logger.Infow("message claimed", "value", string(message.Value), "timestamp", message.Timestamp, "topic", message.Topic)

			var paymentCancelRetryMsg models.PaymentCancelRetryMsg
			err := json.Unmarshal(message.Value, &paymentCancelRetryMsg)
			if err != nil {
				session.MarkMessage(message, "message has invalid body")
				c.logger.Errorw("unmarshal PaymentCancelRetryMsg", "error", err)
				continue
			}

			for time.Now().Sub(message.Timestamp) < time.Second*10 {
			}

			err = c.payment.RetryCancel(session.Context(), paymentCancelRetryMsg.PaymentUid)
			if err != nil {
				c.logger.Warnw("cannot cancel payment", "payment", paymentCancelRetryMsg.PaymentUid, "error", err)
				session.MarkMessage(message, "retry cancel payment")
				c.producer.RetryPaymentCancel(paymentCancelRetryMsg.PaymentUid)
				continue
			}

			c.logger.Infow("cancelled payment", "payment", paymentCancelRetryMsg.PaymentUid)
			session.MarkMessage(message, "got msg for cancel payment")

		case <-session.Context().Done():
			return nil
		}
	}
}
