package main

import (
	"context"
	"jobs/internal/datastore"
	"jobs/scrapers"
	"os"
	"os/signal"
	"sync"
)

func main() {
	var persistentCache = datastore.InitializePersistentCache()
	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer func() {
		persistentCache.CleanUp()
		ctxCancel()
	}()

	var targets []scrapers.Scraper = []scrapers.Scraper{
		&scrapers.Thoughtworks{},
	}
	var wg sync.WaitGroup
	for _, target := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			target.Scrape(ctx)
		}()
	}
	wg.Wait()
}
