package controllers

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
		"SELECT id, name, description, severity FROM alert_types WHERE is_active = TRUE ORDER BY id ASC LIMIT ? OFFSET ?",
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
		if err := rows.Scan(&at.ID, &at.Name, &at.Description, &at.Severity); err != nil {
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
		ncdrURL = "https://alerts.ncdr.nat.gov.tw/api/v1/datastore"
	}

	ncdrKey := os.Getenv("NCDR_API_KEY")
	if ncdrKey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "NCDR API Key not configured",
		})
		return
	}

	// Build query params for NCDR API
	params := url.Values{}
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

	fullURL := ncdrURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

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

	switch format {
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}
