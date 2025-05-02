package reports

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/renniemaharaj/news/internal/types"
)

func reportMatches(r types.Report, q string) bool {
	q = strings.ToLower(q)
	return strings.Contains(strings.ToLower(r.Title), q) ||
		strings.Contains(strings.ToLower(r.Summary), q) ||
		strings.Contains(strings.ToLower(r.URL), q) ||
		strings.Contains(strings.ToLower(r.Date), q) ||
		anyTagMatches(r.Tags, q)
}

func anyTagMatches(tags []string, q string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	re := regexp.MustCompile(`[^\w\-]+`)
	return re.ReplaceAllString(name, "")
}

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		log.Printf("⚠️ Failed to parse date '%s': %v", dateStr, err)
		return time.Time{}
	}
	return t
}
