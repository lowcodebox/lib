package logger

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	// format: "localhost:9092,localhost:9093,localhost:9094"
	Addr              []string
	Topic             string
	NumPartitions     int
	ReplicationFactor int
}

func (kc *KafkaConfig) createTopic() error {
	conn, err := kafka.Dial("tcp", kc.Addr[0])
	if err != nil {
		return errors.Wrap(err, "can't connect to Kafka")
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return errors.Wrap(err, "cannot get controller")
	}

	var controllerConn *kafka.Conn

	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return errors.Wrap(err, "can't connect to Kafka controller")
	}
	defer controllerConn.Close()

	replicationFactor := 1
	if kc.ReplicationFactor > 1 {
		replicationFactor = kc.ReplicationFactor
	}

	numPartitions := 1

	if kc.NumPartitions > 1 {
		replicationFactor = kc.NumPartitions
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             kc.Topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return errors.Wrap(err, "can't create topics")
	}

	return nil
}

func (kc *KafkaConfig) writer(errLogger *zap.Logger) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kc.Addr...),
		Topic:    kc.Topic,
		Balancer: &kafka.LeastBytes{},
		Async:    true,
		ErrorLogger: &errorLogger{
			errLogger,
		},
	}
}

type writerSyncer struct {
	kwr         *kafka.Writer
	errorLogger *zap.Logger
	topic       string
}

func (ws *writerSyncer) Write(p []byte) (int, error) {
	val := make([]byte, len(p))
	copy(val, p)

	m := kafka.Message{
		Value: val,
	}

	err := ws.kwr.WriteMessages(context.Background(), m)
	if err != nil {
		ws.errorLogger.Error("Error writing log", zap.ByteString("log", p), zap.Error(err))
	}

	return len(p), nil
}

func (ws *writerSyncer) Sync() error {
	return ws.kwr.Close()
}

type errorLogger struct {
	*zap.Logger
}

func (l *errorLogger) Printf(msg string, args ...interface{}) {
	l.Error(fmt.Sprintf(msg, args...))
}
