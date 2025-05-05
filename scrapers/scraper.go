package scrapers

type Scraper interface {
	Scrape()
	TargetName() string
}
