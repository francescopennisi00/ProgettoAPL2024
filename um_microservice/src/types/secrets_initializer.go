package types

import (
	"log"
	"os"
)

type SecretInitializer struct{}

var instance *SecretInitializer

func NewSecretInitializer() *SecretInitializer {
	if instance == nil {
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
		log.Fatalf("Error reading secret file for %s: %v", envVarName, err)
	}
	err = os.Setenv(envVarName, string(secretValue))
	if err != nil {
		log.Printf("Initialized %s.\n", envVarName)
		return
	}
}
