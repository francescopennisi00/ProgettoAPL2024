package wms_types

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"time"
	wmsUtils "wms_microservice/wms-src/wms-utils"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(bootstrapServers, acks string) *KafkaProducer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"acks":              acks,
	})
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v\n", err)
	}
	return &KafkaProducer{producer: p}
}

func (kp *KafkaProducer) CreateTopic(broker, topicName string) {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		log.Fatalf("Error creating Kafka admin client: %v\n", err)
	}
	defer admin.Close()

	// retrieve topic's metadata
	topicMetadata, errV := admin.GetMetadata(&topicName, false, wmsUtils.TimeoutTopicMetadata)
	if errV != nil {
		log.Fatalf("Error checking if topic exists: %v", errV)
	}

	// we assume that topic not exists (if it is true we don't enter into for and topicExists remains false)
	topicExists := false
	for _, topic := range topicMetadata.Topics {
		if topic.Topic == topicName {
			topicExists = true
			break
		}
	}

	if !topicExists {
		// create topic if it does not exist
		topicSpecs := []kafka.TopicSpecification{{
			Topic:             topicName,
			NumPartitions:     wmsUtils.KafkaTopicPartitions,
			ReplicationFactor: wmsUtils.KafkaTopicReplicationFactor,
		}}
		results, err := admin.CreateTopics(context.Background(), topicSpecs, nil)
		if err != nil {
			log.Fatalf("Error creating topic %s: %v", topicName, err)
		}
		for _, result := range results {
			fmt.Printf("Topic %s creation error: %s\n", result.Topic, result.Error)
		}

		// Wait for topic creation completion (required in Python, maybe here we can bypass it)
		time.Sleep(3 * time.Second)
	}
}
