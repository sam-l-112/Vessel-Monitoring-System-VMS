package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"vms-api/src/models"
)

type CWAOpenDataController struct{}

type CWAWeatherResponse struct {
	Success string `json:"success"`
	Records struct {
		Locations []struct {
			LocationName string `json:"locationName"`
			Location     []struct {
				LocationName   string `json:"locationName"`
				StationID      string `json:"stationId"`
				WeatherElement []struct {
					ElementName string `json:"elementName"`
					Time        []struct {
						StartTime string `json:"startTime"`
						EndTime   string `json:"endTime"`
						Parameter []struct {
							ParameterName string `json:"parameterName"`
							ParameterUnit string `json:"parameterUnit"`
						} `json:"parameter"`
					} `json:"time"`
				} `json:"weatherElement"`
			} `json:"location"`
		} `json:"locations"`
	} `json:"records"`
}

type CWAWeatherStationResponse struct {
	Success string `json:"success"`
	Records struct {
		Station []struct {
			StationName string `json:"StationName"`
			StationID   string `json:"StationId"`
			ObsTime     struct {
				DateTime string `json:"DateTime"`
			} `json:"ObsTime"`
			GeoInfo struct {
				Coordinates []struct {
					StationLatitude  string `json:"StationLatitude"`
					StationLongitude string `json:"StationLongitude"`
				} `json:"Coordinates"`
				StationAltitude string `json:"StationAltitude"`
				CountyName      string `json:"CountyName"`
				TownName        string `json:"TownName"`
			} `json:"GeoInfo"`
			WeatherElement struct {
				Weather          string `json:"Weather"`
				Precipitation    string `json:"Precipitation"`
				WindDirection    string `json:"WindDirection"`
				WindSpeed        string `json:"WindSpeed"`
				AirTemperature   string `json:"AirTemperature"`
				RelativeHumidity string `json:"RelativeHumidity"`
				AirPressure      string `json:"AirPressure"`
			} `json:"WeatherElement"`
		} `json:"Station"`
	} `json:"records"`
}

func (c *CWAOpenDataController) GetCWAWeatherData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	area := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("area")))
	apiKey := getCWAAPIKey(area)
	if apiKey == "" {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "CWA API Key not configured",
		})
		return
	}

	stationID := r.URL.Query().Get("station")
	stationID = getCWAStationID(stationID, area)

	baseURL := getCWABaseURL()
	apiURL := fmt.Sprintf("%s/O-A0001-001?StationID=%s", baseURL, url.QueryEscape(stationID))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to create request: " + err.Error(),
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

	if resp.StatusCode != http.StatusOK {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("CWA API returned status %d", resp.StatusCode),
		})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to read response body",
		})
		return
	}

	// 嘗試解析新格式 (Station array)
	var stationResp CWAWeatherStationResponse
	if err := json.Unmarshal(body, &stationResp); err == nil {
		if len(stationResp.Records.Station) > 0 {
			weatherData := parseCWAStationData(stationResp, stationID, area)
			json.NewEncoder(w).Encode(models.APIResponse{
				Success: true,
				Data:    weatherData,
				Message: "CWA weather data retrieved successfully",
			})
			return
		}
	}

	// 回退到舊格式
	var weatherResp CWAWeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to parse CWA response: " + err.Error(),
		})
		return
	}

	weatherData := parseCWAWeatherData(weatherResp, stationID, area)

	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    weatherData,
		Message: "CWA weather data retrieved successfully",
	})
}

func (c *CWAOpenDataController) GetCWAForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	area := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("area")))
	apiKey := getCWAAPIKey(area)
	if apiKey == "" {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "CWA API Key not configured",
		})
		return
	}

	locationName := r.URL.Query().Get("location")
	locationName = getCWALocationName(locationName, area)

	forecastType := r.URL.Query().Get("type")
	baseURL := getCWABaseURL()
	var apiURL string

	// 使用不同的 dataset ID
	if area == "newtaipei" {
		// 新北市一周預報
		apiURL = fmt.Sprintf("%s/F-D0047-071?locationName=%s", baseURL, url.QueryEscape(locationName))
	} else if forecastType == "week" {
		// 澎湖一周預報
		apiURL = fmt.Sprintf("%s/F-D0047-047?locationName=%s", baseURL, url.QueryEscape(locationName))
	} else {
		apiURL = fmt.Sprintf("%s/F-C0032-001?locationName=%s", baseURL, url.QueryEscape(locationName))
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to create request: " + err.Error(),
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

	if resp.StatusCode != http.StatusOK {
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("CWA API returned status %d", resp.StatusCode),
		})
		return
	}

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

func getCWABaseURL() string {
	baseURL := os.Getenv("CWA_API_URL")
	if baseURL == "" {
		baseURL = "https://opendata.cwa.gov.tw/api/v1/rest/datastore"
	}
	return strings.TrimRight(baseURL, "/")
}

func getCWAAPIKey(area string) string {
	fmt.Printf("[DEBUG getCWAAPIKey] area=%s\n", area)
	if area == "penghu" {
		key := os.Getenv("CWA_API_KEY_PENGHU")
		fmt.Printf("[DEBUG getCWAAPIKey] CWA_API_KEY_PENGHU=%s\n", key)
		if key != "" {
			return key
		}
	}

	key := os.Getenv("CWA_API_KEY")
	fmt.Printf("[DEBUG getCWAAPIKey] CWA_API_KEY=%s\n", key)
	return key
}

func getCWAStationID(stationID, area string) string {
	if stationID != "" {
		return stationID
	}

	if area == "penghu" {
		return "467350"
	}

	if id := os.Getenv("CWA_STATION_ID"); id != "" {
		return id
	}

	return "466900"
}

func getLocationDisplayName(area string) string {
	if area == "penghu" {
		return "澎湖縣"
	}
	return "新北市"
}

func getCWALocationName(locationName, area string) string {
	if locationName != "" {
		return locationName
	}

	if area == "penghu" {
		return "澎湖縣"
	}

	if name := os.Getenv("CWA_LOCATION_NAME"); name != "" {
		return name
	}

	return "新北市"
}

func parseCWAWeatherData(resp CWAWeatherResponse, stationID string, area string) []models.WeatherData {
	locationName := getLocationDisplayName(area)

	var weatherData []models.WeatherData

	if len(resp.Records.Locations) == 0 {
		return []models.WeatherData{
			{
				ID:              0,
				Temperature:     getDemoTemperature(),
				Humidity:        75.0,
				PhLevel:         7.2,
				DissolvedOxygen: 5.5,
				Location:        locationName + " - " + stationID,
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
				PhLevel:         7.0,
				DissolvedOxygen: windSpeed,
				Location:        location.LocationName + " - " + loc.LocationName,
				RecordedAt:      recordTime,
			})
		}
	}

	if len(weatherData) == 0 {
		weatherData = []models.WeatherData{
			{
				ID:              0,
				Temperature:     getDemoTemperature(),
				Humidity:        75.0,
				PhLevel:         7.2,
				DissolvedOxygen: 5.5,
				Location:        locationName,
				RecordedAt:      time.Now(),
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

func parseCWAStationData(resp CWAWeatherStationResponse, stationID string, area string) []models.WeatherData {
	var weatherData []models.WeatherData

	locationName := getLocationDisplayName(area)

	for _, station := range resp.Records.Station {
		if station.StationID == stationID {
			we := station.WeatherElement

			// 解析溫度
			temp := 0.0
			if we.AirTemperature != "" {
				fmt.Sscanf(we.AirTemperature, "%f", &temp)
			}

			// 解析濕度
			humidity := 0.0
			if we.RelativeHumidity != "" {
				fmt.Sscanf(we.RelativeHumidity, "%f", &humidity)
			}

			// 解析 pH (水質) - 使用預設值或從其他來源
			ph := 7.0

			// 解析溶解氧
			do := 5.5

			// 解析時間
			recordedAt := time.Now()
			if station.ObsTime.DateTime != "" {
				recordedAt, _ = time.Parse("2006-01-02T15:04:05+08:00", station.ObsTime.DateTime)
			}

			weatherData = append(weatherData, models.WeatherData{
				ID:              0,
				Temperature:     temp,
				Humidity:        humidity,
				PhLevel:         ph,
				DissolvedOxygen: do,
				Location:        locationName + " - " + station.StationID,
				RecordedAt:      recordedAt,
			})
			break
		}
	}

	if len(weatherData) == 0 {
		weatherData = []models.WeatherData{
			{
				ID:              0,
				Temperature:     getDemoTemperature(),
				Humidity:        75.0,
				PhLevel:         7.2,
				DissolvedOxygen: 5.5,
				Location:        locationName + " - " + stationID,
				RecordedAt:      time.Now(),
			},
		}
	}

	return weatherData
}
