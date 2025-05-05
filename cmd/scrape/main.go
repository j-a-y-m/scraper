package main

import (
	"jobs/internal/datastore"
	"jobs/scrapers"
)

func main() {
	var persistentCache = datastore.InitializePersistentCache()
	defer persistentCache.CleanUp()

	var targets []scrapers.Scraper = []scrapers.Scraper{
		&scrapers.Thoughtworks{},
	}

	for _, target := range targets {
		target.Scrape()
	}
}
