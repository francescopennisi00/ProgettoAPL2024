package um_utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"psw"`
}

func CalculateHash(inputString string) string {
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(inputString))
	hashResult := sha256Hash.Sum(nil)
	return hex.EncodeToString(hashResult)
}

func SetResponseMessage(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "%s\n", message)
}
