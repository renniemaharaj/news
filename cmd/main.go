package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/renniemaharaj/news/internal/reports"
)

func startHealthPulse(apiURL string) {

	go func() {
		ticker := time.NewTicker(time.Minute / 2)
		defer ticker.Stop()

		client := &http.Client{}

		for range ticker.C {
			if apiURL == "" {
				log.Println("❌ API Address not set for health check")
				continue
			}

			resp, err := client.Get(fmt.Sprintf("%s/healthcheck", apiURL))
			if err != nil {
				log.Printf("❌ health check failed: %v", err)
				continue
			}
			resp.Body.Close()
			log.Printf("✅ health check passed")
		}
	}()
}

func scrapeOnEmptyDir() {
	count := reports.CountReports()
	if count <= 0 {
		reports.ScrapeReports()
	}
}

func main() {
	// Start the report scraping scheduler
	go reports.DailyScheduler()

	// Count reports
	reports.CountReports()

	// scrape on empty dir
	scrapeOnEmptyDir()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	// Setup CORS-wrapped handlers
	handler := reports.CORSMiddleware(http.HandlerFunc(reports.HandleReportRequests))
	http.Handle("/reports", handler)
	http.Handle("/healthcheck", reports.HealthHandler("v1"))

	// Start health pulse
	startHealthPulse(os.Getenv("STAY_ALIVE_API_URL"))

	log.Printf("🟢 API running at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
