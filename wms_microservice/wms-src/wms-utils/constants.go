package wms_utils

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	DBConnString                      = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	KafkaAcksProducerParameter string = strconv.Itoa(int('1'))
)

const (
	UmIpPort                    string = "um-service:50052"
	PortUM                      int    = 50052
	PortAPIGateway              string = "50053"
	DBDriver                    string = "mysql"
	TriggerPeriodTimer                 = 60 * time.Second
	TimeoutTopicMetadata        int    = 5000 // timeout in milliseconds
	KafkaTopicPartitions        int    = 1
	KafkaTopicReplicationFactor int    = 1
	KafkaBootstrapServer        string = "kafka-service:9092"
	KafkaTopicName              string = "event_update"
)
