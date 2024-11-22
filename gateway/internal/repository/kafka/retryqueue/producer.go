package retryqueue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RetryQueueProducer struct {
	producer sarama.AsyncProducer

	carUnbookTopic     string
	paymentCancelTopic string
}

func NewRetryQueueProducer(
	brokers []string,
	carUnbookTopic string,
	paymentCancelTopic string,
	logger *zap.SugaredLogger,
) (*RetryQueueProducer, error) {
	sl, _ := zap.NewStdLogAt(logger.Desugar(), zapcore.WarnLevel)
	sarama.Logger = sl

	config := sarama.NewConfig()
	config.ClientID = "car-rental-system"

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("start kafka async producer: %w", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write access log entry:", err)
		}
	}()

	return &RetryQueueProducer{
		producer:           producer,
		carUnbookTopic:     carUnbookTopic,
		paymentCancelTopic: paymentCancelTopic,
	}, nil
}

func (q *RetryQueueProducer) Stop() {
	q.producer.Close()
}

func (q *RetryQueueProducer) prepareCarUnbookMsg(carUid uuid.UUID) *sarama.ProducerMessage {
	msg := models.CarUnbookRetryMsg{CarUid: carUid}
	marshalledMsg, _ := json.Marshal(msg)
	encoder := sarama.ByteEncoder(marshalledMsg)
	producerMsg := &sarama.ProducerMessage{
		Topic:     q.carUnbookTopic,
		Value:     encoder,
		Timestamp: time.Now(),
	}

	return producerMsg
}

func (q *RetryQueueProducer) RetryCarUnbook(carUid uuid.UUID) {
	q.producer.Input() <- q.prepareCarUnbookMsg(carUid)
}

func (q *RetryQueueProducer) preparePaymentCancelMsg(paymentUid uuid.UUID) *sarama.ProducerMessage {
	msg := models.PaymentCancelRetryMsg{PaymentUid: paymentUid}
	marshalledMsg, _ := json.Marshal(msg)

	encoder := sarama.ByteEncoder(marshalledMsg)
	producerMsg := &sarama.ProducerMessage{
		Topic: q.paymentCancelTopic,
		Value: encoder,
	}

	return producerMsg
}

func (q *RetryQueueProducer) RetryPaymentCancel(paymentUid uuid.UUID) {
	q.producer.Input() <- q.preparePaymentCancelMsg(paymentUid)
}
