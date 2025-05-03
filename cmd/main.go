package main

import (
	"log"
	"net/http"
	"os"

	"github.com/renniemaharaj/news/internal/healthcheck"
	"github.com/renniemaharaj/news/internal/reports"
)

func main() {
	// start the reports scraping scheduler
	go reports.DailyScheduler()

	// running count current will count and log the count
	reports.CountReports()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000" // default port if not set in environment
	}

	// Wrap with CORS middleware
	handler := reports.CORSMiddleware(http.HandlerFunc(reports.HandleReportRequests))

	http.Handle("/reports", handler)
	http.Handle("/healtcheck", reports.HealthHandler("v1"))
	log.Printf("🟢 API running at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	// run the health checker function
	healthcheck.RunHealthCheckFunction()
}
