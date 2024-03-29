package um_types

import (
	"fmt"
	"log"
	"os"
	"sync"
	umUtils "um_microservice/um-src/um-utils"
)

var lock = &sync.Mutex{}

type SecretInitializer struct{}

var instance *SecretInitializer

func NewSecretInitializer() *SecretInitializer {
	if instance == nil {
		lock.Lock() // wait until lock is available
		defer lock.Unlock()
		if instance == nil {
			log.SetPrefix("[INFO] ")
			log.Println("Creating single instance now")
			instance = &SecretInitializer{}
		} else {
			log.SetPrefix("[INFO] ")
			log.Println("Single instance already created")
		}
	} else {
		log.SetPrefix("[INFO] ")
		fmt.Println("Single instance already created")
		instance = &SecretInitializer{}
	}
	return instance
}

func (si *SecretInitializer) InitSecrets() {
	si.initSecret("PASSWORD")
	umUtils.DBConnString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
}

func (si *SecretInitializer) initSecret(envVarName string) {
	secretPath := os.Getenv(envVarName)
	secretValue, err := os.ReadFile(secretPath)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Error reading secret file for %s: %v", envVarName, err)
	}
	err = os.Setenv(envVarName, string(secretValue))
	if err == nil {
		log.SetPrefix("[INFO] ")
		log.Printf("Initialized %s\n", envVarName)
		return
	} else {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Error setting ENV variable %s: %v\n", envVarName, err)
	}
}
