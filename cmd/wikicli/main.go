package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	kafka "github.com/segmentio/kafka-go"
)

const TOPIC = "wikimedia.recentchange"

var (
	urls = []string{"http://localhost:9200"}
)

type MediaWikiEvent struct {
	Title     string `json:"title"`
	Id        int    `json:"id"`
	Type      string `json:"type"`
	Namespace int    `json:"namespace"`
}

func main() {
	start := time.Now()
	msgs := 0

	// make a new reader that consumes from topic-A
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "consumer-group-id",
		Topic:   TOPIC,
		// MinBytes: 1,    // 1
		// MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			time.Sleep(time.Second)
			dur := time.Since(start)
			tput := float64(msgs) / dur.Seconds()

			log.Printf("Messages received %d in %.2f seconds tput %.3f m/s", msgs, dur.Seconds(), tput)
		}
	}()

	// opensearch
	log.Print("Initializing Opensearch client")
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: urls,
		Username:  "admin", // For testing only. Don't store credentials in code.
		Password:  "admin",
	})

	if err != nil {
		log.Fatalf("Error initilizing opensearch client: %v", err)
	}
	log.Print(client.Info())

	for {
		m, err := r.FetchMessage(context.Background())
		// m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Error reading message: %v", err)
			break
		}

		ev := MediaWikiEvent{}
		err = json.Unmarshal(m.Value, &ev)
		if err != nil {
			log.Fatalf("could not unmarshall event %v: %v", string(m.Value), err)
		}

		// Very bruteforce
		value := string(m.Value)
		for char := range "+-&|!(){}[]^\"~*?:\\" {
			value = strings.ReplaceAll(
				value,
				fmt.Sprintf("%c", char),
				fmt.Sprintf("\\%c", char),
			)
		}
		// value = strings.ReplaceAll(value, "[", "\\[")
		// value = strings.ReplaceAll(value, "]", "\\]")
		// value = strings.ReplaceAll(value, "]", "\\]")
		// + - && || ! ( ) { } [ ] ^ " ~ * ? : \

		req := opensearchapi.IndexRequest{
			Index:      "wikimedia",
			DocumentID: strconv.Itoa(ev.Id),
			Body:       strings.NewReader(string(value)),
		}
		insertResponse, err := req.Do(context.Background(), client)
		if err != nil {
			log.Fatalf("could not instert document %s, response %v: %v", m.Value, insertResponse, err)
		}

		// Commit offset on Kafka
		err = r.CommitMessages(context.Background(), m)
		if err != nil {
			log.Printf("Error: Could not commit message")
		}

		// fmt.Printf("message at %v topic/partition/offset %v/%v/%v: %s = %s\n", m.Time, m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		msgs++
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
