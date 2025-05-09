package scrapers

import "context"

type Scraper interface {
	Scrape(context.Context)
	TargetName() string
}
