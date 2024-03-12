package wms_types

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"time"
	wmsUtils "wms_microservice/wms-src/wms-utils"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(bootstrapServers string, acks string) *KafkaProducer {
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

// CreateTopic create topic if not yet exists
func (kp *KafkaProducer) CreateTopic(broker string, topicName string) {
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
		log.Fatalf("Error checking if topic exists: %v\n", errV)
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
		// create topic if it does not exist as kafka package requires
		topicSpecs := []kafka.TopicSpecification{{
			Topic:             topicName,
			NumPartitions:     wmsUtils.KafkaTopicPartitions,
			ReplicationFactor: wmsUtils.KafkaTopicReplicationFactor,
		}}
		results, err := admin.CreateTopics(context.Background(), topicSpecs, nil)
		if err != nil {
			log.SetPrefix("[ERROR] ")
			log.Fatalf("Error creating topic %s: %v\n", topicName, err)
		}
		for _, result := range results {
			log.SetPrefix("[INFO] ")
			log.Printf("Topic %s creation error: %v\n", result.Topic, result.Error)
		}

		// Wait for topic creation completion (required in Python, maybe here we can bypass it)
		time.Sleep(3 * time.Second)
	}
}

// Optional per-message delivery callback when a message has been successfully
// delivered or permanently failed delivery.
// Updating table user_constraints in order to avoid considering again a row in the building of
// Kafka message to publish in "event_update" topic. In this way, we prevent multiple replication of
// the WMS from sending the same trigger message to worker
func deliveryCallback(ack *kafka.Message) error {

	if ack.TopicPartition.Error != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Delivery failed: %v\n", ack.TopicPartition.Error)
		return ack.TopicPartition.Error
	}

	// ack contains message, that is a JSON formatted string received as []byte: we convert it into a KafkaMessage type
	var message wmsUtils.KafkaMessage
	if err := json.Unmarshal(ack.Value, &message); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error decoding message: %v\n", err)
		return err
	}
	log.SetPrefix("[INFO] ")
	log.Printf("Ack received for message: ")
	log.Println(message)

	rowsIDList := message.RowsIdList

	// open DB connection in order to update timestamp of the last check for the rules of the ack message
	var dbConn DatabaseConnector
	_, err := dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB connection error! -> %v\n", err)
		return err
	}

	// we have to commit changes only after the conclusion of the for. So, we start a DB transaction
	// that we don't commit until the end of the for
	_, errTr := dbConn.BeginTransaction()
	if errTr != nil {
		log.SetPrefix("ERROR ")
		log.Printf("Error in start DB transaction: %v\n", errTr)
		return errTr
	}
	for _, id := range rowsIDList {
		// id is an integer but its type is string
		query := fmt.Sprintf("UPDATE user_constraints SET time_stamp = CURRENT_TIMESTAMP() WHERE id = %s", id)
		_, _, err := dbConn.ExecIntoTransaction(query)
		if err != nil {
			log.SetPrefix("[ERROR] ")
			log.Printf("Error executing update query: %v\n", err)
			_ = dbConn.RollbackTransaction()
			return err
		}
	}
	// now we can commit the transaction
	errCom := dbConn.CommitTransaction()
	if errCom != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error in commit transaction: %v\n", errCom)
		return errCom
	}

	// if we arrived here all went well
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

	// wait for ack from Kafka broker
	log.SetPrefix("[INFO] ")
	log.Println("Waiting for message to be delivered")
	event := <-deliveryChan
	ack := event.(*kafka.Message)

	return deliveryCallback(ack)
}

func (*KafkaProducer) MakeKafkaMessage(locationId string) (string, error) {

	// fetch location information from the database
	var dbConn DatabaseConnector
	_, err := dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB connection error! -> %v\n", err)
		return "", err
	}
	// extracting from DB the list of information about current location of the Kafka message
	query := fmt.Sprintf("SELECT location_name, latitude, longitude, country_code, state_code FROM locations WHERE id = %s", locationId)
	_, locationRows, errQ := dbConn.ExecuteQuery(query)
	if errQ != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error during DB select for fetching location info: %v\n", errQ)
		return "", errQ
	}

	// fetch user constraints from the database
	query = fmt.Sprintf("SELECT * FROM user_constraints WHERE TIMESTAMPDIFF(SECOND,  time_stamp, CURRENT_TIMESTAMP()) > trigger_period AND location_id = %s", locationId)
	_, userConstraintsRows, errV := dbConn.ExecuteQuery(query)
	if errV != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error during DB select for fetching user constraints info: %v\n", errV)
		return "", errV
	}

	var kMessage wmsUtils.KafkaMessage
	for _, userConstraintsRow := range userConstraintsRows {

		var rulesJson wmsUtils.RulesIntoDB
		// we put 'rules' field (it is a json into DB row in third position) into rulesJson
		if err := json.Unmarshal([]byte(userConstraintsRow[3]), &rulesJson); err != nil {
			log.SetPrefix("[ERROR] ")
			log.Printf("Error during unmarshaling of rules JSON from DB into KafkaMessage type: %v\n", err)
			return "", err
		}

		kMessage.UserIdList = append(kMessage.UserIdList, userConstraintsRow[1])
		kMessage.RowsIdList = append(kMessage.RowsIdList, userConstraintsRow[0])
		kMessage.MaxTempList = append(kMessage.MaxTempList, rulesJson.MaxTemp)
		kMessage.MinTempList = append(kMessage.MinTempList, rulesJson.MinTemp)
		kMessage.MaxHumidityList = append(kMessage.MaxHumidityList, rulesJson.MaxHumidity)
		kMessage.MinHumidityList = append(kMessage.MinHumidityList, rulesJson.MinHumidity)
		kMessage.MaxPressureList = append(kMessage.MaxPressureList, rulesJson.MaxPressure)
		kMessage.MinPressureList = append(kMessage.MinPressureList, rulesJson.MinPressure)
		kMessage.MaxCloudList = append(kMessage.MaxCloudList, rulesJson.MaxCloud)
		kMessage.MinCloudList = append(kMessage.MinCloudList, rulesJson.MinCloud)
		kMessage.MaxWindSpeedList = append(kMessage.MaxWindSpeedList, rulesJson.MaxWindSpeed)
		kMessage.MinWindSpeedList = append(kMessage.MinWindSpeedList, rulesJson.MinWindSpeed)
		kMessage.WindDirectionList = append(kMessage.WindDirectionList, rulesJson.WindDirection)
		kMessage.RainList = append(kMessage.RainList, rulesJson.Rain)
		kMessage.SnowList = append(kMessage.SnowList, rulesJson.Snow)
	}

	kMessage.Location = locationRows[0] //the Kafka Message has only one location that is the only row returned

	finalJsonBytes, errMar := json.Marshal(kMessage)
	if errMar != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error during marshaling location info into []byte: %v\n", errMar)
		return "", errMar
	}

	log.Println("FINAL JSON DICT: ", string(finalJsonBytes))
	return string(finalJsonBytes), nil

}
