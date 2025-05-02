package types

type Report struct {
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	Tags      []string `json:"tags"`
	URL       string   `json:"url"`
	Date      string   `json:"date"`
	Relevance int      `json:"relevance"`
}
