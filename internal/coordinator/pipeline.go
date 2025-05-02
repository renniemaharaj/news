package coordinator

import (
	"fmt"

	"github.com/renniemaharaj/news/internal/browser"
	"github.com/renniemaharaj/news/internal/config"
	"github.com/renniemaharaj/news/internal/model"
	"github.com/renniemaharaj/news/internal/types"
)

// Reports channel
var Channel = make(chan types.Report, 100)

// Coordinator runner
func Run(cfg *config.Config) error {
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
			Channel <- r
			fmt.Printf("✔️ [%s] %s\n", r.Tags, r.Title)
		}
	}
	return nil
}
