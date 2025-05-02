package reports

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Request handler
func HandleReportRequests(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	index, _ := strconv.Atoi(r.URL.Query().Get("index"))
	max, _ := strconv.Atoi(r.URL.Query().Get("max"))
	relStr := r.URL.Query().Get("relevance")
	desiredRelevance, _ := strconv.Atoi(relStr)

	if max <= 0 {
		max = 10
	}

	reports, err := loadReports(query, index, max, desiredRelevance)
	if err != nil {
		http.Error(w, "Failed to load reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}
