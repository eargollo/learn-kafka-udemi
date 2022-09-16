package main

import (
	"context"
	"log"
	"time"

	"github.com/donovanhide/eventsource"
	"github.com/segmentio/kafka-go"
)

const TOPIC = "wikimedia.recentchange"

func main() {
	start := time.Now()
	msgs := 0

	// event stream reader
	stream, err := eventsource.Subscribe("https://stream.wikimedia.org/v2/stream/recentchange", "")
	if err != nil {
		log.Fatalf("Could not subscribe to source: %v", err)
	}

	// kafka producer
	w := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        TOPIC,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
		// BatchTimeout: 300 * time.Microsecond,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Printf("ERROR: error sending messages: %v", err)
			}
		},
		// BatchTimeout: time.Minute,
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

	// Thread to care for errors at the stream
	go func() {
		for {
			err := <-stream.Errors
			log.Printf("ERROR on incoming stream: %v", err)
		}
	}()

	for {
		ev := <-stream.Events
		// fmt.Printf("%d: Id %s Event [%s]\n", msgs, ev.Id(), ev.Event()) //, ev.Data())
		err = w.WriteMessages(context.Background(),
			kafka.Message{
				// Key:   []byte("Key-A"),
				Value: []byte(ev.Data()),
			},
		)
		if err != nil {
			log.Printf("Error writing to kafka stream: %v", err)
			break
		}
		msgs++
	}

}
