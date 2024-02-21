package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	pb "um_microservice"
	si "um_microservice/src/types"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedNotifierUmServer
}

func (s *server) RequestEmail(ctx context.Context, in *pb.Request) (*pb.Reply, error) {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE")))
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.SetPrefix("[ERROR]")
			log.Println("DB connection closing error!")
		}
	}(db)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error connecting to the database: %v", err)
		return &pb.Reply{Email: "null"}, nil
	}

	row := db.QueryRow("SELECT email FROM users WHERE id=?", int(in.UserId))
	var email string
	err = row.Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pb.Reply{Email: "not present anymore"}, nil
		} else {
			log.SetPrefix("[ERROR] ")
			log.Printf("Error scanning row: %v\n", err)
			return &pb.Reply{Email: "null"}, err
		}
	} else {
		log.SetPrefix("[INFO] ")
		log.Printf("Returning email %s\n to Notifier", email)
		return &pb.Reply{Email: email}, nil
	}

}

func serveNotifier() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen to requests from Notifier: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterNotifierUmServer(s, &server{})
	log.Printf("Notifier server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve Notifier: %v", err)
	}
}

func main() {

	siInstance := si.NewSecretInitializer()
	siInstance.InitSecrets()
}
