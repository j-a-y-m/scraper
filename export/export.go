package export

type JobPosting struct {
	Company        string
	Role           string
	Qualifications string
	Location       string
	Description    string
	Id             string
	Url            string
}

type ScrapingError struct {
	Company string
	Url     string
	Err     error
}

func (e *ScrapingError) Error() string {
	return e.Company + ": " + e.Url + ": " + e.Err.Error()
}

func (e *ScrapingError) Unwrap() error {
	return e.Err
}
