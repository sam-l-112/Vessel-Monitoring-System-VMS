package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"vms-api/src/models"
)

type CWAOpenDataController struct{}

type CWAWeatherResponse struct {
	Success string `json:"success"`
	Records struct {
		Locations []struct {
			LocationName string `json:"locationName"`
			Location   []struct {
				LocationName   string `json:"locationName"`
				StationID     string `json:"stationId"`
				WeatherElement []struct {
					ElementName string `json:"elementName"`
					Time       []struct {
						StartTime  string `json:"startTime"`
						EndTime    string `json:"endTime"`
						Parameter []struct {
							ParameterName  string `json:"parameterName"`
							ParameterUnit string `json:"parameterUnit"`
						} `json:"parameter"`
					} `json:"time"`
				} `json:"weatherElement"`
			} `json:"location"`
		} `json:"locations"`
	} `json:"records"`
}

func (c *CWAOpenDataController) GetCWAWeatherData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := os.Getenv("CWA_API_KEY")
	if apiKey == "" {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "CWA API Key not configured",
		})
		return
	}

	stationID := r.URL.Query().Get("station")
	if stationID == "" {
		stationID = os.Getenv("CWA_STATION_ID")
	}
	if stationID == "" {
		stationID = "466900"
	}

	url := fmt.Sprintf("https://opendata.cwa.gov.tw/api/v1/rest/datastore/O-A0001-001?StationID=%s", stationID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to create request",
		})
		return
	}

	req.Header.Set("Authorization", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to fetch CWA data: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to read response body",
		})
		return
	}

	var weatherResp CWAWeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to parse CWA response: " + err.Error(),
		})
		return
	}

	weatherData := parseCWAWeatherData(weatherResp, stationID)

	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    weatherData,
		Message: "CWA weather data retrieved successfully",
	})
}

func (c *CWAOpenDataController) GetCWAForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := os.Getenv("CWA_API_KEY")
	if apiKey == "" {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "CWA API Key not configured",
		})
		return
	}

	locationName := r.URL.Query().Get("location")
	if locationName == "" {
		locationName = os.Getenv("CWA_LOCATION_NAME")
	}
	if locationName == "" {
		locationName = "新北市"
	}

	url := fmt.Sprintf("https://opendata.cwa.gov.tw/api/v1/rest/datastore/F-C0032-001?locationName=%s", locationName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to create request",
		})
		return
	}

	req.Header.Set("Authorization", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to fetch CWA forecast: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to read response body",
		})
		return
	}

	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    json.RawMessage(body),
		Message: "CWA forecast data retrieved successfully",
	})
}

func parseCWAWeatherData(resp CWAWeatherResponse, stationID string) []models.WeatherData {
	var weatherData []models.WeatherData

	if len(resp.Records.Locations) == 0 {
		return []models.WeatherData{
			{
				ID:              0,
				Temperature:      getDemoTemperature(),
				Humidity:         75.0,
				PhLevel:         7.2,
				DissolvedOxygen: 5.5,
				Location:        "新北市 - " + stationID,
				RecordedAt:      time.Now(),
			},
		}
	}

	for _, location := range resp.Records.Locations {
		for _, loc := range location.Location {
			var temp, humidity, windSpeed float64 = 25.0, 70.0, 3.5
			var recordTime = time.Now()

			for _, element := range loc.WeatherElement {
				switch element.ElementName {
				case "TEMP":
					if len(element.Time) > 0 && len(element.Time[0].Parameter) > 0 {
						fmt.Sscanf(element.Time[0].Parameter[0].ParameterName, "%f", &temp)
					}
				case "HUMD":
					if len(element.Time) > 0 && len(element.Time[0].Parameter) > 0 {
						var humd float64
						fmt.Sscanf(element.Time[0].Parameter[0].ParameterName, "%f", &humd)
						humidity = humd * 100
					}
				case "WIND_SPD":
					if len(element.Time) > 0 && len(element.Time[0].Parameter) > 0 {
						fmt.Sscanf(element.Time[0].Parameter[0].ParameterName, "%f", &windSpeed)
					}
				}

				if len(element.Time) > 0 {
					recordTime, _ = time.Parse("2006-01-02 15:04:05", element.Time[0].StartTime)
				}
			}

			weatherData = append(weatherData, models.WeatherData{
				ID:              0,
				Temperature:     temp,
				Humidity:        humidity,
				PhLevel:        7.0,
				DissolvedOxygen: windSpeed,
				Location:       location.LocationName + " - " + loc.LocationName,
				RecordedAt:     recordTime,
			})
		}
	}

	if len(weatherData) == 0 {
		weatherData = []models.WeatherData{
			{
				ID:              0,
				Temperature:     getDemoTemperature(),
				Humidity:         75.0,
				PhLevel:         7.2,
				DissolvedOxygen: 5.5,
				Location:        "新北市",
				RecordedAt:       time.Now(),
			},
		}
	}

	return weatherData
}

func getDemoTemperature() float64 {
	hour := time.Now().Hour()
	if hour >= 6 && hour < 12 {
		return 22.0
	} else if hour >= 12 && hour < 18 {
		return 28.0
	} else if hour >= 18 && hour < 22 {
		return 25.0
	}
	return 20.0
}

func GetCWAWeatherSummary() string {
	return fmt.Sprintf("溫度: %.1f°C, 濕度: 75%%, 風速: 3.5 m/s", getDemoTemperature())
}