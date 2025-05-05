package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// HealthHandler responds to healthcheck requests
func HealthHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "OK %s", version)
	}
}

// Request handler
func HandleReportRequests(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	index, _ := strconv.Atoi(r.URL.Query().Get("index"))
	max, _ := strconv.Atoi(r.URL.Query().Get("max"))
	relStr := r.URL.Query().Get("relevance")
	desiredRelevance, _ := strconv.Atoi(relStr)

	if max <= 0 {
		max = 30
	}

	reports, err := loadReports(query, index, max, desiredRelevance)
	if err != nil {
		http.Error(w, "Failed to load reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}
