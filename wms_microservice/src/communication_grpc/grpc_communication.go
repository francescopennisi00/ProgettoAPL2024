package communication_grpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strings"
	protoBuf "wms_microservice/proto"
	"wms_microservice/src/types"
)

type WmsUmServer struct {
	protoBuf.UnimplementedWMSUmServer
}

func (s *WmsUmServer) RequestDeleteUserConstraints(ctx context.Context, in *protoBuf.User) (*protoBuf.ResponseCode, error) {

	// extracting user id from the request
	userId := in.GetUserId()

	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err := dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB connection error! -> %v\n", err)
	}
	query := fmt.Sprintf("SELECT * FROM user_constraints WHERE user_id= %d", userId)
	_, _, errV := dbConn.ExecuteQuery(query)
	if errors.Is(errV, sql.ErrNoRows) {
		return &protoBuf.ResponseCode{Code: 0}, nil // the user has no constraints
	}
	query = fmt.Sprintf("DELETE FROM user_constraints WHERE user_id= %d", userId)
	_, _, errV = dbConn.ExecuteQuery(query)
	if errV != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error in DB while deleting detected user constraints: %v\n", errV)
		return &protoBuf.ResponseCode{Code: -1}, nil // something went wrong in DB execution query
	}
	return &protoBuf.ResponseCode{Code: 0}, nil // delete completed!

}

func AuthenticateAndRetrieveUserId(authorizationHeader string) (int64, error) {

	jwtToken := strings.Split(authorizationHeader, " ")[1]

	// start gRPC communication with User Manager in order to retrieve user id
	conn, errV := grpc.Dial("um-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)
	if errV != nil {
		return 0, errV
	}
	client := protoBuf.NewWMSUmClient(conn)

	response, errVar := client.RequestUserIdViaJWTToken(context.Background(), &protoBuf.Request{JwtToken: jwtToken})
	if errVar != nil {
		return 0, errVar
	}

	// no gRPC error
	return response.UserId, nil // user id < 0 if some error occurred at server

}
