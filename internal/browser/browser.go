package browser

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func resolveURL(link string, base string) string {
	uri, err := url.Parse(link)
	if err != nil {
		return link
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return link
	}
	return baseURL.ResolveReference(uri).String()
}

func isLikelyThumbnail(src string) bool {
	lower := strings.ToLower(src)
	// Avoid logos, icons, placeholders, and SVGs
	return !strings.Contains(lower, "logo") &&
		!strings.Contains(lower, "icon") &&
		!strings.Contains(lower, "svg") &&
		!strings.Contains(lower, "placeholder") &&
		(strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".webp"))
}

// Browser scraping method, returns string of all textContent + images
func Scrape(url string) (string, error) {
	log.Printf("🗃️ Visiting site for scraping: %s", url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	// Scrape all <p> tags
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			sb.WriteString(text + "\n")
		}
	})

	// Collect image URLs
	var images []string
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && isLikelyThumbnail(src) {
			images = append(images, resolveURL(src, url))
		}
	})

	// Append the images at the bottom
	if len(images) > 0 {
		sb.WriteString("\n\nimages:\n")
		for _, img := range images {
			sb.WriteString(img + "\n")
		}
	}

	sb.WriteString(fmt.Sprintf("\n\nsource_url=%s", url))
	return sb.String(), nil
}

var skipDomains = map[string]struct{}{
	"maps.google.com":   {},
	"photos.google.com": {},
}

// Browser searching method, return search results
func Search(query string, numSitesPerQuery int) ([]string, error) {
	searchURL := "https://www.google.com/search?q=" + strings.ReplaceAll(query, " ", "+") + "&num=20&tbm=nws"
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	var results []string

	for {
		if len(results) >= numSitesPerQuery {
			break
		}

		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return results, nil
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, attr := range t.Attr {
					if attr.Key == "href" && strings.HasPrefix(attr.Val, "/url?q=") {
						extracted := strings.Split(attr.Val[7:], "&")[0]
						if strings.HasPrefix(extracted, "http") {
							u, err := url.Parse(extracted)
							if err != nil {
								continue
							}
							domain := u.Hostname()
							if _, skip := skipDomains[domain]; skip {
								continue
							}

							// Check for 404 before including
							respCheck, err := http.Head(extracted)
							if err != nil || respCheck.StatusCode == http.StatusNotFound {
								log.Printf("⚠️ Skipping 404 result: %s", u)
								continue
							}

							results = append(results, strings.TrimSpace(extracted))
							break
						}
					}
				}
			}
		}
	}

	return results, nil
}
