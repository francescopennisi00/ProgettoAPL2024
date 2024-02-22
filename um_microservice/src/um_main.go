package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	pb "um_microservice"
	"um_microservice/src/types"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedNotifierUmServer
}

func (s *server) RequestEmail(ctx context.Context, in *pb.Request) (*pb.Reply, error) {

	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err := dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		err := database.CloseConnection()
		if err != nil {
			log.SetPrefix("[ERROR] ")
			log.Println("Error in closing DB connection!")
		}
	}(&dbConn)
	if err != nil {
		return &pb.Reply{Email: "null"}, nil
	}

	var email interface{}
	userId := in.UserId
	query := fmt.Sprintf("SELECT email FROM users WHERE id=%d", userId)
	_, e := dbConn.ExecuteQuery(true, query, &email)
	if e != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Printf("Email not present anymore")
			return &pb.Reply{Email: "not present anymore"}, nil
		} else {
			log.SetPrefix("[ERROR] ")
			log.Printf("DB Error: %v\n", err)
			return &pb.Reply{Email: "null"}, err
		}
	} else {
		if emailString, ok := email.(string); ok {
			return &pb.Reply{Email: emailString}, nil
		} else {
			log.Println("Email is not a string!")
			return &pb.Reply{Email: "null"}, nil
		}
	}
}

func serveNotifier() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Fatalf("Failed to listen to requests from Notifier: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterNotifierUmServer(s, &server{})
	log.SetPrefix("[INFO] ")
	log.Printf("Notifier server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve Notifier: %v", err)
	}
}

func main() {

	siInstance := types.NewSecretInitializer()
	siInstance.InitSecrets()

	log.Println("ENV variables initialization done!")
}
