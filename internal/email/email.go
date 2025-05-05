package email

import (
	"errors"
	"net/mail"
)

type EmailClient interface {
	initialize(config any) EmailClient
	SendMail(to []string, subject, body string) error
	SenderAddress() mail.Address
}

var errClientUnInitialized = errors.New("mail client needs to be initialise before usage")
