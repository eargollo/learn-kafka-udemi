package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	start := time.Now()
	msgs := 0

	// make a writer that produces to topic-A, using the least-bytes distribution
	w := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        "topic-A",
		Async:        true,
		BatchTimeout: 300 * time.Microsecond,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Printf("ERROR: error sending messages: %v", err)
			}
		},
		// Balancer: &kafka.LeastBytes{},
	}

	defer w.Close()

	go func() {
		for {
			time.Sleep(time.Second)
			dur := time.Since(start)
			tput := float64(msgs) / dur.Seconds()

			log.Printf("Messages sent %d in %.2f seconds tput %.3f m/s", msgs, dur.Seconds(), tput)
		}
	}()

	for {
		err := w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(strconv.Itoa(msgs)),
				Value: []byte(fmt.Sprintf("Message number %d", msgs)),
			},
		)
		randomSleep()
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
		msgs++
	}

}

func randomSleep() {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(10) // n will be between 0 and 10
	// fmt.Printf("Sleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Millisecond)
	// fmt.Println("Done")
}
