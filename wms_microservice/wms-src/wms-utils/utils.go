package wms_utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
)

type Credentials struct {
	Email    string
	Password string
}

type Rules struct {
	TriggerPeriod string
	LocationInfo  []string
	MaxTemp       string
	MinTemp       string
	MaxHumidity   string
	MinHumidity   string
	MaxPressure   string
	MinPressure   string
	MaxWindSpeed  string
	MInWindSpeed  string
	WindDirection string
	Rain          string
	Snow          string
	MaxCloud      string
	MinCloud      string
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

func Round(x float64, places int) float64 {
	shift := math.Pow10(places)
	return math.Round(x*shift) / shift
}
