package wms_utils

import (
	"fmt"
	"os"
	"time"
)

var (
	DBConnString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
)

const (
	UmIpPort           string = "um-service:50052"
	PortUM             int    = 50052
	PortAPIGateway     string = "50053"
	DBDriver           string = "mysql"
	TriggerPeriodTimer        = 60 * time.Second
)
