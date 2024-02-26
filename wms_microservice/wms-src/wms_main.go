package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
	wmsUm "wms_microservice/wms-proto"
	grpcC "wms_microservice/wms-src/wms-communication_grpc"
	httpC "wms_microservice/wms-src/wms-http_handlers"
	wmsTypes "wms_microservice/wms-src/wms-types"
	wmsUtils "wms_microservice/wms-src/wms-utils"
)

var wg sync.WaitGroup

var (
	portUm = flag.Int("portUM", wmsUtils.PortUM, "The server port for UM")
)

func timer(duration time.Duration, event chan<- bool) {
	defer wg.Done()
	for {
		//every duration the timer wakes up the goroutine waiting trigger period in order to send update
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

func serveAPIGateway() {
	defer wg.Done()
	port := wmsUtils.PortAPIGateway

	hostname, _ := os.Hostname()
	log.SetPrefix("[INFO] ")
	log.Printf("Hostname: %s server starting on port: %s\n", hostname, port)

	router := mux.NewRouter()
	router.HandleFunc("/update_rules", httpC.UpdateRulesHandler).Methods("POST")
	router.HandleFunc("/show_rules", httpC.ShowRulesHandler).Methods("GET")
	router.HandleFunc("/update_rules/delete_user_constraints_by_location'", httpC.DeleteRulesHandler).Methods("POST")
	log.SetPrefix("[ERROR] ")
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func findPendingWork(kProducer *wmsTypes.KafkaProducer) ([]string, error) {
	var dbConn wmsTypes.DatabaseConnector
	_, err := dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *wmsTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB connection error! -> %v\n", err)
		return nil, err
	}
	query := "SELECT location_id FROM user_constraints WHERE TIMESTAMPDIFF(SECOND,  time_stamp, CURRENT_TIMESTAMP()) > trigger_period GROUP BY location_id"
	_, rows, err := dbConn.ExecuteQuery(query)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB query execution error in retreive location ids! -> %v\n", err)
		return nil, err
	}
	var kafkaMessageList []string
	for _, row := range rows {
		locationId := row[0]
		kafkaMessage, err := kProducer.MakeKafkaMessage(locationId)
		if err != nil {
			log.SetPrefix("[ERROR] ")
			log.Printf("Error in MakeKafkaMessage -> %v\n", err)
			return nil, err
		}
		kafkaMessageList = append(kafkaMessageList, kafkaMessage)
	}
	return kafkaMessageList, nil
}

func triggerPeriodGoRoutine(expiredTimer chan bool, producer *wmsTypes.KafkaProducer) {
	defer wg.Done()
	for {
		// check in DB in order to find updates to send
		kafkaMessageList, err := findPendingWork(producer)
		if err != nil {
			log.SetPrefix("[ERROR] ")
			log.Fatalf("Error in findPendingWork! -> %v\n", err)
		}
		for _, message := range kafkaMessageList {
			for {
				err := producer.ProduceKafkaMessage(wmsUtils.KafkaTopicName, message)
				if err == nil {
					break
				}
			}
		}
		// wait for expired timer event
		<-expiredTimer
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

	log.SetPrefix("[INFO] ")
	log.Println("Starting serving API gateway goroutine!")
	go serveAPIGateway()
	log.SetPrefix("[INFO] ")
	log.Println("Starting serving UM goroutine!")
	go serveUm()
	log.Println("Starting timer goroutine!")
	expiredTimerEvent := make(chan bool)
	go timer(wmsUtils.TriggerPeriodTimer, expiredTimerEvent)

	// go routine made in order to wait trigger period
	go triggerPeriodGoRoutine(expiredTimerEvent, kafkaProducer)

	wg.Add(4)
	wg.Wait()
	log.SetPrefix("[INFO] ")
	log.Println("All goroutines have finished. Exiting...")

}
