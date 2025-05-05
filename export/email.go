package export

import (
	"jobs/internal/email"
	"log"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type Email struct{}

func (*Email) Export(exports chan any) error {

	var emailClient email.EmailClient = email.GetGmailClient()
	emailRecipientsEnv := os.Getenv("EMAIL_RECIPIENTS")

	for res := range exports {
		emailRecipients := strings.Split(emailRecipientsEnv, ",")
		if len(emailRecipients) < 1 {
			log.Println("env variable EMAIL_RECIPIENTS invalid")
			break
		}
		if jobPosting, ok := res.(JobPosting); ok {
			err := emailJob(emailRecipients, jobPosting, emailClient)
			if err != nil {
				log.Println(err)
			}
		} else if err, ok := res.(ScrapingError); ok {
			err := emailError(emailRecipients, err, emailClient)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return nil
}

func emailError(emailRecipients []string, err ScrapingError, emailClient email.EmailClient) error {
	subject := "Scraping error: " + err.Company
	body := "Error occured while scraping : " + "[url]" + "(" + err.Url + ")" + "\n " +
		"error: \n " + err.Err.Error()
	mailErr := emailClient.SendMail(emailRecipients, subject, body)
	return mailErr
}

func emailJob(emailRecipients []string, jp JobPosting, emailClient email.EmailClient) error {
	subject := jp.Company + " - " + jp.Role + " (" + jp.Location + ")"
	body := "<b>Qualifications:</b> " + jp.Qualifications + " <br> " +
		"<hr>" +
		"<b>Description:</b> " + jp.Description + " <br> " +
		"<hr>" +
		"<b>url:</b> " + jp.Url + " <br> "

	mailErr := emailClient.SendMail(emailRecipients, subject, body)

	return mailErr
}
