package wms_types

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var lock = &sync.Mutex{}

type SecretInitializer struct{}

var instance *SecretInitializer

func NewSecretInitializer() *SecretInitializer {
	if instance == nil {
		lock.Lock()
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
}

func (si *SecretInitializer) initSecret(envVarName string) {
	secretPath := os.Getenv(envVarName)
	secretValue, err := os.ReadFile(secretPath)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Error reading secret file for %s: %v", envVarName, err)
	}
	err = os.Setenv(envVarName, string(secretValue))
	if err != nil {
		log.SetPrefix("[INFO] ")
		log.Printf("Initialized %s\n", envVarName)
		return
	}
}
