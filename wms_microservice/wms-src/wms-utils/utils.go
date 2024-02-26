package wms_utils

import (
	"fmt"
	"math"
	"net/http"
)

type Credentials struct {
	Email    string
	Password string
}

type Rules struct {
	TriggerPeriod string   `json:"trigger_period"`
	LocationInfo  []string `json:"location"`
	MaxTemp       string   `json:"max_temp"`
	MinTemp       string   `json:"min_temp"`
	MaxHumidity   string   `json:"max_humidity"`
	MinHumidity   string   `json:"min_humidity"`
	MaxPressure   string   `json:"max_pressure"`
	MinPressure   string   `json:"min_pressure"`
	MaxWindSpeed  string   `json:"max_wind_speed"`
	MinWindSpeed  string   `json:"min_wind_speed"`
	WindDirection string   `json:"wind_direction"`
	Rain          string   `json:"rain"`
	Snow          string   `json:"snow"`
	MaxCloud      string   `json:"max_cloud"`
	MinCloud      string   `json:"min_cloud"`
}

type RulesIntoDB struct {
	MaxTemp       string `json:"max_temp"`
	MinTemp       string `json:"min_temp"`
	MaxHumidity   string `json:"max_humidity"`
	MinHumidity   string `json:"min_humidity"`
	MaxPressure   string `json:"max_pressure"`
	MinPressure   string `json:"min_pressure"`
	MaxWindSpeed  string `json:"max_wind_speed"`
	MinWindSpeed  string `json:"min_wind_speed"`
	WindDirection string `json:"wind_direction"`
	Rain          string `json:"rain"`
	Snow          string `json:"snow"`
	MaxCloud      string `json:"max_cloud"`
	MinCloud      string `json:"min_cloud"`
}

type ShowRulesOutput struct {
	Rules         RulesIntoDB `json:"rules"`
	Location      []string    `json:"location"`
	TriggerPeriod string      `json:"trigger_period"`
}

type Location []string

type KafkaMessage struct {
	UserIdList        []string `json:"user_id"`
	Location          []string `json:"location"`
	MaxTempList       []string `json:"max_temp"`
	MinTempList       []string `json:"min_temp"`
	MaxHumidityList   []string `json:"max_humidity"`
	MinHumidityList   []string `json:"min_humidity"`
	MaxPressureList   []string `json:"max_pressure"`
	MinPressureList   []string `json:"min_pressure"`
	MaxWindSpeedList  []string `json:"max_wind_speed"`
	MinWindSpeedList  []string `json:"min_wind_speed"`
	WindDirectionList []string `json:"wind_direction"`
	RainList          []string `json:"rain"`
	SnowList          []string `json:"snow"`
	MaxCloudList      []string `json:"max_cloud"`
	MinCloudList      []string `json:"min_cloud"`
	RowsIdList        []string `json:"rows_id"`
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
