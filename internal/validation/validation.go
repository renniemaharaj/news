package validation

import (
	"encoding/json"
	"fmt"

	"github.com/renniemaharaj/news/internal/types"
)

type ReportWrapper struct {
	Reports []types.Report
}

func Validate(resp string) error {
	var wrapper ReportWrapper
	err := json.Unmarshal([]byte(resp), &wrapper)
	if err != nil {
		return err
	}

	for i, report := range wrapper.Reports {
		if report.Title == "" || report.Summary == "" {
			return fmt.Errorf("⚠️ report %d has missing fields", i)
		}
		if report.Relevance < 1 || report.Relevance > 10 {
			return fmt.Errorf("⚠️ report %d has invalid relevance: %d", i, report.Relevance)
		}
	}

	return nil
}
