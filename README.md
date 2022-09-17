# Learn Kafka

Exercices from Udemi Kafka course, done in Golang instead of Java

## Dev environment

Kafka and Zookeper are run as a docker compose with data serialized at `data`

```
$ docker-compose up
```

OpenSearch is another docker compose also serializing to `data`

```
$ docker-compose -f open-serach-compose.yaml up
```

Cleaning up the environment completely

```
$ docker-compose rm
$ docker-compose -f open-serach-compose.yaml rm
$ sudo rm -rf data
```

## Examples

### Simple producer consumer
Creating a Kafka topic and adding one or more producers/consumers. 

Exploring different producers and consumers options.

Creating topic and explore the effect on different partitions configurations (out of order for example):
```
$ kafka-topics.sh --bootstrap-server localhost:9092 --create --topic topic-A --partitions 3
```

Describe topics:
```
$ kafka-topics.sh --bootstrap-server localhost:9092 --describe
```

In one terminal run consumer:
```
$ go run cmd/consumer/main.go 
```

In another terminal run producer:
```
$ go run cmd/producer/main.go
```

At first attempt I noticed that producer only sent 1 message per second. 
Solved issue by making producer assynchronous and got 1.5 million produced messages per second. 

Bottleneck became the consumer that was receiving a message at a time with a throughput of 500 messages per second. Could not get the throughput beyond 800 messages per second even by adding 10k bytes of minimum for receiving messages or increasing the queue capacity. Strange.

Only solution might be having multiple clients. Added a random sleep up to 10ms to control the producer to about 200 messages/s.

### Wikimedia to OpenSearch

Created Wikimedia producer reading from online stream. Same pattern. Avg tput from stream is 14 mes/sec.

Put up OpenSeach. Check:
```
# Get status
curl -XGET --insecure -u 'admin:admin' 'http://localhost:9200'
```

Creating index:
```
curl -XPUT --insecure -u 'admin:admin' 'http://localhost:9200/wikimedia'
```

Got idepotency by unmarshalling and using the event id. Could use the document id as well.

To check if there is imbalance between producer and consumer, we can query the consumer group:
```
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --group consumer-group-id --describe
```

Changed from read message to fetch message and then made client reliable with an offset commit to Kafka only after adding the record to OpenSearch.
