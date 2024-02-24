package utils

import (
	"fmt"
	"os"
)

var (
	DBConnString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
)

const (
	WmsIpPort      string = "wms-service:50052"
	PortWMS        int    = 50052
	PortNotifier   int    = 50051
	PortAPIGateway string = "50053"

	HourTokenExpiration int = 72
)
