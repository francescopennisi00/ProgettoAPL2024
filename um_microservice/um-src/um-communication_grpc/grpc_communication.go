package um_communication_grpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strconv"
	notifierUm "um_microservice/um-proto/notifier_um"
	wmsUm "um_microservice/um-proto/wms_um"
	umTypes "um_microservice/um-src/um-types"
	umUtils "um_microservice/um-src/um-utils"
)

type UmNotifierServer struct {
	notifierUm.UnimplementedNotifierUmServer
}

type UmWmsServer struct {
	wmsUm.UnimplementedWMSUmServer
}

func DeleteUserConstraintsByUserId(userId string) error {

	//convert userId from string to int64
	num, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		fmt.Printf("Error during conversion: %v\n", err)
		return err
	}

	// start connection to gRPC server
	conn, errV := grpc.Dial(umUtils.WmsIpPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)
	if errV != nil {
		return errV
	}
	client := wmsUm.NewWMSUmClient(conn)

	response, errVar := client.RequestDeleteUserConstraints(context.Background(), &wmsUm.User{UserId: num})
	if errVar != nil {
		return errVar
	}
	if response.Code != 0 {
		errVariable := errors.New("something went wrong in WMS while deleting user constraints")
		return errVariable
	}

	return nil //no error occurred
}

func (s *UmWmsServer) RequestUserIdViaJWTToken(ctx context.Context, in *wmsUm.Request) (*wmsUm.Reply, error) {

	// extracting JWT token from the request
	tokenString := in.GetJwtToken()

	// extracting token information without verifying them: needed in order to retrieve user email
	token, _ := jwt.Parse(tokenString, nil)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error in extracting JWT Token without verifying it with Claims")
		err := errors.New("error in extracting JWT Token without verifying it with Claims")
		return nil, err
	}
	email, okBool := claims["email"].(string)
	if !okBool {
		log.SetPrefix("[ERROR] ")
		log.Printf("Impossible to extract email from JWT Token without verifying it")
		err := errors.New("impossible to extract email from JWT Token without verifying it")
		return nil, err
	}

	// Retrieve user's password from DB in oder to verifying JWT Token and authenticate him
	var dbConn umTypes.DatabaseConnector
	_, err := dbConn.StartDBConnection(umUtils.DBConnString)
	defer func(database *umTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB connection error! -> %v\n", err)
		return nil, err
	}
	query := fmt.Sprintf("SELECT id, password FROM users WHERE email= '%s'", email)
	_, row, errV := dbConn.ExecuteQuery(query)
	if errV != nil {
		if errors.Is(errV, sql.ErrNoRows) {
			return &wmsUm.Reply{UserId: -3}, nil // token is not valid: email not present
		}
		log.SetPrefix("[ERROR] ")
		log.Printf("DB query error! -> %v\n", errV)
		return nil, errV
	}

	//convert user id from string to int64
	var idUser int64
	idUser, err = strconv.ParseInt(row[0][0], 10, 64)
	if err != nil {
		fmt.Printf("Error: user id returned by DB is not an integer: -> %v\n", err)
		return nil, err
	}

	//password is already a string: no conversion is needed
	password := row[0][1]

	// verify JWT Token with password as secret
	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(password), nil
	})
	if errors.Is(err, jwt.ErrTokenExpired) {
		return &wmsUm.Reply{UserId: -1}, nil //token is expired
	}
	if !token.Valid {
		return &wmsUm.Reply{UserId: -3}, nil //token is not valid: password incorrect
	}

	// Everything succeeded, return user ID
	return &wmsUm.Reply{UserId: idUser}, nil

}

func (s *UmNotifierServer) RequestEmail(ctx context.Context, in *notifierUm.Request) (*notifierUm.Reply, error) {

	var dbConn umTypes.DatabaseConnector
	_, err := dbConn.StartDBConnection(umUtils.DBConnString)
	defer func(database *umTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		return &notifierUm.Reply{Email: "null"}, nil
	}

	userId := in.UserId
	query := fmt.Sprintf("SELECT email FROM users WHERE id=%d", userId)
	_, emailRows, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Println("Email not present anymore")
			return &notifierUm.Reply{Email: "not present anymore"}, nil
		} else {
			log.SetPrefix("[ERROR] ")
			log.Printf("DB Error: %v\n", errorVar)
			return &notifierUm.Reply{Email: "null"}, nil
		}
	} else {
		emailString := emailRows[0][0]
		return &notifierUm.Reply{Email: emailString}, nil
	}
}
