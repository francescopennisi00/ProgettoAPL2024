package main

import (
	"log"
	"os"
)

var logger *log.Logger

func initLogger() {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	logger.SetPrefix("[INFO] ")
	logger.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	initLogger()

}
