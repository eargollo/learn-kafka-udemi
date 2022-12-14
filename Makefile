KAFKA_PATH=~/tools/kafka_2.13-3.2.1

.PHONY:run
run:
	docker-compose up

.PHONY: run-nativr
run-native: zookeeper-daemon kafka

.PHONY: lint
lint: lint-code lint-security lint-vulnerability

.PHONY: lint-code
lint-code:
	golangci-lint run ./...

.PHONY: lint-security
lint-security:
	gosec ./...

.PHONY: lint-vulnerability
lint-vulnerability:
	govulncheck ./...

.PHONY: list
list:
	kafka-topics.sh --list --bootstrap-server eduardos-air.local:9092

.PHONY: kafka
kafka:
	${KAFKA_PATH}/bin/kafka-server-start.sh ./config/server.properties

.PHONY: zookeeper
zookeeper:
	${KAFKA_PATH}/bin/zookeeper-server-start.sh ./config/zookeeper.properties 

.PHONY: zookeeper-daemon
zookeeper-daemon:
	${KAFKA_PATH}/bin/zookeeper-server-start.sh -daemon ./config/zookeeper.properties 

.PHONY: zookeeper-stop
zookeeper-stop:
	${KAFKA_PATH}/bin/zookeeper-server-stop.sh