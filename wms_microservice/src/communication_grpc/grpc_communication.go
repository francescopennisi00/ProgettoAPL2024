package communication_grpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strings"
	protoBuf "wms_microservice/proto"
	wmsTypes "wms_microservice/src/types"
	wmsUtils "wms_microservice/src/utils"
)

type WmsUmServer struct {
	protoBuf.UnimplementedWMSUmServer
}

func (s *WmsUmServer) RequestDeleteUserConstraints(ctx context.Context, in *protoBuf.User) (*protoBuf.ResponseCode, error) {

	// extracting user id from the request
	userId := in.GetUserId()

	var dbConn wmsTypes.DatabaseConnector
	_, err := dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *wmsTypes.DatabaseConnector) {
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
	conn, errV := grpc.Dial(wmsUtils.UmIpPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
