package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

//https://github.com/googleworkspace/go-samples/blob/main/gmail/quickstart/quickstart.go

type GmailClient struct {
	gmailService  *gmail.Service
	senderAddress *mail.Address
}

var gmailClient EmailClient

func (client *GmailClient) initialize(conf any) EmailClient {
	//TODO add mutex check while init
	ctx := context.Background()
	b, err := os.ReadFile("cred.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailMetadataScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	gmailClient := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(gmailClient))

	client.gmailService = srv

	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	client.senderAddress = client.getSenderAddress()
	return client
}

func (client *GmailClient) SendMail(to []string, subject, body string) error {
	client.checkGmailService()
	_, encodedMsg := createAndEncodeEmail(client.SenderAddress(), to, subject, body)

	_, err := client.gmailService.Users.Messages.Send("me", &gmail.Message{
		Raw: encodedMsg,
	}).Do()

	return err
}

func (client *GmailClient) SenderAddress() mail.Address {
	if client.senderAddress == nil {
		client.senderAddress = client.getSenderAddress()
	}

	return *client.senderAddress
}

func (client *GmailClient) getSenderAddress() *mail.Address {
	client.checkGmailService()
	usersGetProfileCall := client.gmailService.Users.GetProfile("me")
	profile, err := usersGetProfileCall.Do()
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client user profile: %v", err)
	}

	return &mail.Address{
		Address: profile.EmailAddress,
		Name:    "jobscraper",
	}
}

func GetGmailClient() EmailClient {
	if gmailClient == nil {
		gmailClient = &GmailClient{}
		gmailClient.initialize(nil)
		return gmailClient
	} else {
		return gmailClient
	}
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (client *GmailClient) checkGmailService() {
	if client.gmailService == nil {
		log.Fatalln(fmt.Errorf("gmail client: %w", errClientUnInitialized))
	}
}

func createAndEncodeEmail(from mail.Address, to []string, subject, body string) (string, string) {
	var msgBuilder strings.Builder
	date := time.Now().Format(time.RFC1123Z)

	msgBuilder.WriteString(fmt.Sprintf("From: %s\r\n", from.String()))
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", date))

	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n")

	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(body)

	rawMessage := msgBuilder.String()
	fmt.Println(rawMessage)

	encodedMessage := base64.URLEncoding.EncodeToString([]byte(rawMessage))
	return rawMessage, encodedMessage
}
