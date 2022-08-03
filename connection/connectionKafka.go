package connection

import (
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

// ConnectionKafka create connection with kafka
func ConnectionKafka() *kafka.Reader {
	kafkaConnection := os.Getenv("KAFKA_CONNECTION")
	config := kafka.ReaderConfig{
		Brokers:  []string{kafkaConnection},
		GroupID:  "consumer-scraping",
		Topic:    "topic-scraping",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	}

	fmt.Println("Connection Apache Kafka")

	return kafka.NewReader(config)
}
