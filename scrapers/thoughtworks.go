package scrapers

import (
	"context"
	"encoding/json"
	jobsExport "jobs/export"
	"jobs/internal/datastore"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/geziyor/geziyor/export"

	"github.com/go-playground/validator/v10"
)

const (
	orgName     = "Thoughtworks"
	listingsUrl = "https://www.thoughtworks.com/rest/careers/jobs"
	detailUrl   = "https://www.thoughtworks.com/en-in/careers/jobs/"
)

type Thoughtworks struct {
}

func (*Thoughtworks) TargetName() string {
	return orgName
}

func (*Thoughtworks) Scrape(ctx context.Context) {

	scraperOptions := geziyor.Options{
		URLRevisitEnabled: true,
		StartURLs:         []string{listingsUrl},
		ParseFunc:         scrapeJobListings,
		ErrorFunc: func(g *geziyor.Geziyor, r *client.Request, err error) {
			g.Exports <- jobsExport.ScrapingError{
				Company: orgName,
				Url:     r.RequestURI,
				Err:     err,
			}
		},
		Exporters: []export.Exporter{
			&export.PrettyPrint{},
			&jobsExport.Email{},
		},
	}

	pollDuration := 5 * time.Minute
	ticker := time.NewTicker(pollDuration)
	for {
		scraperOptionsCopy := scraperOptions
		scraper := geziyor.NewGeziyor(&scraperOptionsCopy)
		scraper.Start()
		ticker.Reset(pollDuration)
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

type JobListings struct {
	Jobs []Job `validate:"required"`
}

type Job struct {
	Country        string `validate:"required"`
	Name           string `validate:"required"`
	SourceSystemId int    `validate:"required"`
}

func scrapeJobListings(g *geziyor.Geziyor, r *client.Response) {
	var jobListings JobListings
	parseError := json.Unmarshal(r.Body, &jobListings)

	if parseError == nil {
		parseError = validator.New().Struct(jobListings)
	}

	if parseError != nil {

		g.Exports <- jobsExport.ScrapingError{
			Company: orgName,
			Url:     listingsUrl,
			Err:     parseError,
		}
		return

	}

	for _, job := range jobListings.Jobs {
		if isJobRelevant(job) {
			scrapJobDetails(g, job)
		}
	}

}

func scrapJobDetails(g *geziyor.Geziyor, job Job) {
	var cache datastore.Cache = datastore.GetPersistentCache()

	id := job.SourceSystemId
	jdUrl := detailUrl + strconv.Itoa(id)
	val, _ := cache.Get(orgName, strconv.Itoa(id))
	if v, ok := val.([]byte); ok && string(v) == "" {
		g.Get(jdUrl, func(g *geziyor.Geziyor, r *client.Response) {

			jdJson := strings.Join(strings.Fields(r.HTMLDoc.Find(`script[type="application/ld+json"]`).Text()), " ")
			var jd struct {
				Title            string `validate:"required"`
				Description      string
				Responsibilities string
				Qualifications   string
				Skills           string
				JobLocation      struct {
					Address struct {
						AddressLocality string `validate:"required"`
					} `validate:"required"`
				} `validate:"required"`
			}
			parseError := json.Unmarshal([]byte(jdJson), &jd)

			if parseError == nil {
				parseError = validator.New().Struct(jd)
			}

			if parseError != nil {

				g.Exports <- jobsExport.ScrapingError{
					Company: orgName,
					Url:     jdUrl,
					Err:     parseError,
				}
				return

			}

			jp := jobsExport.JobPosting{
				Company:        orgName,
				Role:           jd.Title,
				Description:    jd.Description + "/n" + "responsibilities: " + jd.Responsibilities,
				Location:       jd.JobLocation.Address.AddressLocality,
				Qualifications: jd.Qualifications + "/n" + jd.Skills,
				Id:             strconv.Itoa(id),
				Url:            jdUrl,
			}

			g.Exports <- jp
			cache.Put(orgName, strconv.Itoa(id), time.Now().Format(time.RFC822))
		})
	} else {
		log.Println("skipping "+jdUrl+" already processed at ", val)
	}

}

func isJobRelevant(job Job) bool {
	country := job.Country
	role := job.Name
	mustContainReg, mustContainRegEr := regexp.Compile("(?i)(software development|developer)")
	mustNotContainReg, mustNotContainRegEr := regexp.Compile("(?i)(senior|lead)")
	if mustContainRegEr != nil || mustNotContainRegEr != nil {
		log.Println(mustContainRegEr, mustNotContainRegEr)
		return false
	}
	return country == "India" && mustContainReg.MatchString(role) && !mustNotContainReg.MatchString(role)
}
