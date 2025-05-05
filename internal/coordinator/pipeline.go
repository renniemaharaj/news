package coordinator

import (
	"fmt"

	"github.com/renniemaharaj/news/internal/browser"
	"github.com/renniemaharaj/news/internal/config"
	"github.com/renniemaharaj/news/internal/model"
	"github.com/renniemaharaj/news/internal/types"
)

// Coordinator runner
func Run(cfg *config.Config, output chan types.Report) error {
	defer close(output) // Only coordinator closes it after sending

	for _, keyword := range cfg.Keywords {
		fmt.Println("🔍 Google searching: ", keyword)
		urls, err := browser.Search(keyword, cfg.NumSitesPerQuery)
		if err != nil {
			return err
		}

		var textContents []string
		for _, url := range urls {

			content, err := browser.Scrape(url)
			if err == nil {
				textContents = append(textContents, content)
			}
		}

		reportWrapper, err := model.Prompt(textContents)
		if err != nil {
			return err
		}
		reports := reportWrapper.Reports

		for _, r := range reports {
			output <- r
			fmt.Printf("✔️ [%s] %s\n", r.Tags, r.Title)
		}
	}
	return nil
}
