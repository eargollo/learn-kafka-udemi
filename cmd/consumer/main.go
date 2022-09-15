package main

import (
	"context"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// const TOPIC = "wikimedia.recentchange"
const TOPIC = "topic-A"

func main() {
	start := time.Now()
	msgs := 0

	// make a new reader that consumes from topic-A
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:       []string{"localhost:9092"},
		GroupID:       "consumer-group-id",
		Topic:         TOPIC,
		MinBytes:      10e5, // 1KB
		MaxBytes:      10e6, // 10MB
		QueueCapacity: 1000,
	})

	go func() {
		for {
			time.Sleep(time.Second)
			dur := time.Since(start)
			tput := float64(msgs) / dur.Seconds()

			log.Printf("Messages received %d in %.2f seconds tput %.3f m/s", msgs, dur.Seconds(), tput)
		}
	}()

	for {
		_, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Error reading message: %v", err)
			break
		}
		// fmt.Printf("message at %v topic/partition/offset %v/%v/%v: %s = %s\n", m.Time, m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		msgs++
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}

}
