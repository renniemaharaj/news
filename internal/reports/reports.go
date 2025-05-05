package reports

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/renniemaharaj/news/internal/config"
	"github.com/renniemaharaj/news/internal/coordinator"

	"github.com/renniemaharaj/news/internal/types"
)

const (
	reportsDir       = "./reports"
	reportingHour    = 8
	reportExpiration = 72 * time.Hour // 3 days
)

// Counts current reports and logs
func CountReports() int {
	reports, err := loadReports("", 0, 0, 0)
	if err != nil {
		log.Println(fmt.Errorf("error loading reports: %v", err))
	}
	log.Printf("📊 Found %d reports", len(reports))

	return len(reports)
}

// The daily report-scraping scheduler
func DailyScheduler() {
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), reportingHour, 0, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		log.Printf("⌛ Next scraping scheduled for: %s", nextRun.Format(time.RFC1123))
		time.Sleep(time.Until(nextRun))
		ScrapeReports()
	}
}

// Daily report-scraper scraper function
func ScrapeReports() {
	channel := make(chan types.Report)

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("⚠️ Failed to load config: %s", err)
		return
	}

	// Save goroutine reads reports
	go func() {
		for report := range channel {
			saveReport(report)
		}
	}()

	if err := os.MkdirAll(reportsDir, os.ModePerm); err != nil {
		log.Printf("⚠️ Failed to create reports directory: %s", err)
	}

	cleanExpiredReports()

	// Run coordinator pipeline (it will close the channel when done)
	if err := coordinator.Run(cfg, channel); err != nil {
		log.Printf("⚠️ Pipeline error: %s", err)
	}
}

// SaveReport function saves the report to reports directory
func saveReport(report types.Report) {
	safeTitle := sanitizeFilename(report.Title)
	filename := filepath.Join(reportsDir, fmt.Sprintf("%s.json", safeTitle))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("⚠️ Failed to marshal report %s: %v", report.Title, err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("⚠️ Failed to write file %s: %v", filename, err)
		return
	}

	log.Printf("✔️ Report saved: %s", filename)
}

// Local cleanExpiredReports function loadsReports and removes expired ones
func cleanExpiredReports() {
	reports, err := loadReports("", 0, 0, -1) // Load all reports regardless of relevance
	if err != nil {
		log.Printf("⚠️ Failed to load reports: %v", err)
		return
	}

	now := time.Now()
	for _, report := range reports {
		reportTime := parseDate(report.Date)
		if reportTime.IsZero() {
			log.Printf("⚠️ Invalid date in report %s: skipping expiration check", report.Title)
			continue
		}

		if now.Sub(reportTime) > reportExpiration {
			filename := filepath.Join(reportsDir, fmt.Sprintf("%s.json", sanitizeFilename(report.Title)))
			if err := os.Remove(filename); err != nil {
				log.Printf("⚠️ Failed to delete expired report %s: %v", filename, err)
			} else {
				log.Printf("⚠️ Expired report removed: %s", filename)
			}
		}
	}
}

func loadReports(search string, index, max, desiredRelevance int) ([]types.Report, error) {
	var matched []types.Report

	err := filepath.WalkDir(reportsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var report types.Report
		if err := json.Unmarshal(data, &report); err != nil {
			log.Printf("⚠️  Failed to unmarshal %s: %v", path, err)
			return nil
		}

		parsed := parseDate(report.Date)
		if report.Date == "" || parsed.IsZero() {
			now := time.Now().UTC().Format(time.RFC3339)
			log.Printf("⚠️ Found missing or invalid date in %s. Updating to current time: %s", report.Title, now)
			report.Date = now
			saveReport(report)
		}

		if (search == "" || reportMatches(report, search)) && (report.Relevance > desiredRelevance) {
			matched = append(matched, report)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.SliceStable(matched, func(i, j int) bool {
		di, dj := parseDate(matched[i].Date), parseDate(matched[j].Date)

		if di.IsZero() && dj.IsZero() {
			return false
		}
		if di.IsZero() {
			return false
		}
		if dj.IsZero() {
			return true
		}

		if di.Equal(dj) {
			return matched[i].Relevance > matched[j].Relevance
		}
		return di.After(dj)
	})

	if index >= len(matched) {
		log.Printf("⚠️ Requested index %d exceeds matched reports (%d)", index, len(matched))
		return []types.Report{}, nil
	}

	end := index + max
	if max == 0 || end > len(matched) {
		end = len(matched)
	}

	return matched[index:end], nil
}
