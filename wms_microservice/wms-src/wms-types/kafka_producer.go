package wms_types

import (
	"context"
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
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Error creating Kafka producer: %v\n", err)
	}
	return &KafkaProducer{producer: p}
}

func (kp *KafkaProducer) CreateTopic(broker, topicName string) {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Error creating Kafka admin client: %v\n", err)
	}
	defer admin.Close()

	// retrieve topic's metadata
	topicMetadata, errV := admin.GetMetadata(&topicName, false, wmsUtils.TimeoutTopicMetadata)
	if errV != nil {
		log.SetPrefix("[ERROR] ")
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
			log.SetPrefix("[ERROR] ")
			log.Fatalf("Error creating topic %s: %v", topicName, err)
		}
		for _, result := range results {
			log.SetPrefix("[INFO] ")
			log.Printf("Topic %s creation error: %s\n", result.Topic, result.Error)
		}

		// Wait for topic creation completion (required in Python, maybe here we can bypass it)
		time.Sleep(3 * time.Second)
	}
}

// Optional per-message delivery callback when a message has been successfully
// delivered or permanently failed delivery.
func deliveryCallback(ack *kafka.Message) error {
	//TODO implement this!!!
	return nil
}

func (kp *KafkaProducer) ProduceKafkaMessage(topicName string, message string) error {
	// publish on the specific topic
	deliveryChan := make(chan kafka.Event)
	err := kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, deliveryChan)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error producing Kafka message: %v\n", err)
		return err
	}
	// wait ack from Kafka broker
	log.SetPrefix("[INFO] ")
	log.Println("Waiting for message to be delivered")
	event := <-deliveryChan
	ack := event.(*kafka.Message)
	return deliveryCallback(ack)
}
