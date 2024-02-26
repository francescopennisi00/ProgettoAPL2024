package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"time"
	wmsUm "wms_microservice/wms-proto"
	grpcC "wms_microservice/wms-src/wms-communication_grpc"
	wmsTypes "wms_microservice/wms-src/wms-types"
	wmsUtils "wms_microservice/wms-src/wms-utils"
)

var wg sync.WaitGroup

var (
	portUm = flag.Int("portUM", wmsUtils.PortUM, "The server port for UM")
)

func timer(duration time.Duration, event chan<- bool) {
	for {
		time.Sleep(duration)
		event <- true
	}
}

func serveUm() {
	defer wg.Done()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *portUm))
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to listen to requests from Um: %v\n", err)
	}
	wmsServer := grpc.NewServer()
	wmsUm.RegisterWMSUmServer(wmsServer, &grpcC.WmsUmServer{})
	log.SetPrefix("[INFO] ")
	log.Printf("UM server listening at %v\n", lis.Addr())
	if err := wmsServer.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve UM: %v\n", err)
	}
}

func main() {

	siInstance := wmsTypes.NewSecretInitializer()
	siInstance.InitSecrets()

	log.SetPrefix("[INFO] ")
	log.Println("ENV variables initialization done!")

	// Creating table 'location' and 'user_constraints' if not exist
	var dbConn wmsTypes.DatabaseConnector
	_, err := dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *wmsTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB connection error! -> %v\n", err)
	}
	query := "CREATE TABLE IF NOT EXISTS locations (id INTEGER PRIMARY KEY AUTO_INCREMENT, location_name VARCHAR(100) NOT NULL, latitude FLOAT NOT NULL, longitude FLOAT NOT NULL, country_code VARCHAR(10) NOT NULL, state_code VARCHAR(70) NOT NULL, UNIQUE KEY location_tuple (location_name, latitude, longitude));"
	_, _, err = dbConn.ExecuteQuery(query)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB error in creating 'locations' table: %v\n", err)
	}
	query = "CREATE TABLE IF NOT EXISTS user_constraints (id INTEGER PRIMARY KEY AUTO_INCREMENT, user_id INTEGER NOT NULL, location_id INTEGER NOT NULL, rules JSON NOT NULL, time_stamp TIMESTAMP NOT NULL, trigger_period INTEGER NOT NULL, checked BOOLEAN NOT NULL DEFAULT FALSE, FOREIGN KEY (location_id) REFERENCES locations(id), UNIQUE KEY user_location_id (user_id, location_id));"
	_, _, err = dbConn.ExecuteQuery(query)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB error in creating 'user_constraints' table: %v\n", err)
	}

	// create Kafka topic on which to publish (if not yet exists)
	kafkaProducer := wmsTypes.NewKafkaProducer(wmsUtils.KafkaBootstrapServer, wmsUtils.KafkaAcksProducerParameter)
	kafkaProducer.CreateTopic(wmsUtils.KafkaBootstrapServer, wmsUtils.KafkaTopicName)

	//TODO check in DB in order to find events to send

	log.SetPrefix("[INFO] ")
	log.Println("Starting serving API gateway goroutine!")
	// TODO APIGateway?
	// go serveAPIGateway()
	log.SetPrefix("[INFO] ")
	log.Println("Starting serving UM goroutine!")
	go serveUm()

	log.Println("Starting timer goroutine!")
	expiredTimerEvent := make(chan bool)
	go timer(wmsUtils.TriggerPeriodTimer, expiredTimerEvent)

	for {
		// Wait for expired timer event
		<-expiredTimerEvent

		// TODO Check in DB to find events to send

	}
	// Codice inraggiungibile lasciamo perdere wg.wait() oppure usiamo un'altra goroutine per
	// fare check in DB to find events to send?
	wg.Add(1)
	wg.Wait()
	log.SetPrefix("[INFO] ")
	log.Println("All goroutines have finished. Exiting...")

}
