package controllers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"vms-api/src/database"
	"vms-api/src/models"
)

type alertTypesXMLResponse struct {
	XMLName xml.Name          `xml:"response"`
	Success bool              `xml:"success"`
	Message string            `xml:"message"`
	Data    []models.AlertType `xml:"data>alert_type"`
	Total   int               `xml:"total"`
}

func GetAlertTypes(w http.ResponseWriter, r *http.Request) {
	apikey := r.URL.Query().Get("apikey")
	format := r.URL.Query().Get("format")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if apikey == "" || format == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Missing required parameters: apikey, format",
		})
		return
	}

	expectedKey := os.Getenv("OPENCLAW_API_KEY")
	if expectedKey == "" {
		expectedKey = "openclaw_vms_secret_key_2026"
	}
	if apikey != expectedKey {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Invalid API key",
		})
		return
	}

	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	offset := 0
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM alert_types WHERE is_active = TRUE").Scan(&total)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to query alert types",
		})
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, name, description, severity, is_active FROM alert_types WHERE is_active = TRUE ORDER BY id ASC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to query alert types",
		})
		return
	}
	defer rows.Close()

	var alertTypes []models.AlertType
	for rows.Next() {
		var at models.AlertType
		if err := rows.Scan(&at.ID, &at.Name, &at.Description, &at.Severity, &at.IsActive); err != nil {
			continue
		}
		alertTypes = append(alertTypes, at)
	}

	switch format {
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(alertTypesXMLResponse{
			Success: true,
			Message: "操作成功",
			Data:    alertTypes,
			Total:   total,
		})
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: true,
			Message: "操作成功",
			Data: map[string]interface{}{
				"alert_types": alertTypes,
				"total":       total,
			},
		})
	}
}

func GetAlertMessages(w http.ResponseWriter, r *http.Request) {
	apikey := r.URL.Query().Get("apikey")
	format := r.URL.Query().Get("format")

	if apikey == "" || format == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Missing required parameters: apikey, format",
		})
		return
	}

	expectedKey := os.Getenv("OPENCLAW_API_KEY")
	if expectedKey == "" {
		expectedKey = "openclaw_vms_secret_key_2026"
	}
	if apikey != expectedKey {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Invalid API key",
		})
		return
	}

	ncdrURL := os.Getenv("NCDR_API_URL")
	if ncdrURL == "" {
		ncdrURL = "https://alerts.ncdr.nat.gov.tw/api/datastore"
	}

	ncdrKey := os.Getenv("NCDR_API_KEY")
	ncdrKey = strings.TrimSpace(ncdrKey)
	if ncdrKey == "" {
		// NCDR key 未設定時回傳空清單，不報錯
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":"true","result":{"count":"0","items":[]}}`))
		return
	}

	// NCDR API: apikey 必須用 query parameter (非 Authorization header)
	params := url.Values{}
	params.Set("apikey", ncdrKey)
	params.Set("format", "json")
	params.Set("limit", "50")
	params.Set("offset", "0")

	if v := r.URL.Query().Get("capcode"); v != "" {
		params.Set("capcode", v)
	}
	if v := r.URL.Query().Get("govcode"); v != "" {
		params.Set("govcode", v)
	}
	if v := r.URL.Query().Get("countycode"); v != "" {
		params.Set("countycode", v)
	}
	if v := r.URL.Query().Get("effectivetime"); v != "" {
		params.Set("effectivetime", v)
	}
	if v := r.URL.Query().Get("expirestime"); v != "" {
		params.Set("expirestime", v)
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		params.Set("limit", v)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		params.Set("offset", v)
	}

	fullURL := ncdrURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to create NCDR request: " + err.Error(),
		})
		return
	}

	req.Header.Set("Authorization", ncdrKey)
	req.Header.Set("x-api-key", ncdrKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "VMS-Aquaculture/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to connect to NCDR API: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to read NCDR response",
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("NCDR API %s: %s", resp.Status, string(body)),
		})
		return
	}

	switch format {
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func GetCWAAlerts(w http.ResponseWriter, r *http.Request) {
	apikey := r.URL.Query().Get("apikey")
	dataset := r.URL.Query().Get("dataset")

	if apikey == "" || dataset == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Missing required parameters: apikey, dataset",
		})
		return
	}

	expectedKey := os.Getenv("OPENCLAW_API_KEY")
	if expectedKey == "" {
		expectedKey = "openclaw_vms_secret_key_2026"
	}
	if apikey != expectedKey {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Invalid API key",
		})
		return
	}

	cwaKey := os.Getenv("CWA_API_KEY")
	if cwaKey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "CWA API Key not configured. Please set CWA_API_KEY in .env",
		})
		return
	}

	cwaBaseURL := os.Getenv("CWA_API_URL")
	if cwaBaseURL == "" {
		cwaBaseURL = "https://opendata.cwa.gov.tw/api/v1/rest/datastore"
	}
	cwaBaseURL = strings.TrimRight(cwaBaseURL, "/")

	params := url.Values{}
	params.Set("Authorization", cwaKey)
	params.Set("format", "json")

	if v := r.URL.Query().Get("limit"); v != "" {
		params.Set("limit", v)
	} else {
		params.Set("limit", "20")
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		params.Set("offset", v)
	}

	for _, key := range []string{"AreaName", "StationName", "StationID", "CountyName", "TownName",
		"geocode", "severity_level", "expires", "sort", "timeFrom", "timeTo", "WeatherElement"} {
		if v := r.URL.Query().Get(key); v != "" {
			params.Set(key, v)
		}
	}

	fullURL := fmt.Sprintf("%s/%s?%s", cwaBaseURL, dataset, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to create CWA request: " + err.Error(),
		})
		return
	}

	req.Header.Set("Authorization", cwaKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to connect to CWA API: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Failed to read CWA response",
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("CWA API %s: %s", resp.Status, string(body)),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
