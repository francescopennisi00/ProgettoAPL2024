package wms_utils

import (
	"strconv"
	"time"
)

var (
	DBConnString               string
	KafkaAcksProducerParameter = strconv.Itoa(1)
)

const (
	UmIpPort                    string = "um-service:50052"
	PortUM                      int    = 50052
	PortAPIGateway              string = "50051"
	DBDriver                    string = "mysql"
	TriggerPeriodTimer                 = 60 * time.Second
	TimeoutTopicMetadata        int    = 5000 // timeout in milliseconds
	KafkaTopicPartitions        int    = 1
	KafkaTopicReplicationFactor int    = 1
	KafkaBootstrapServer        string = "kafka-service:9092"
	KafkaTopicName              string = "event_update"
)
