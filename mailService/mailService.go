package mailService

import (
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type MailParams struct {
	To               MailUser
	From             MailUser
	Subject          string
	PlainTextContent string
	HTMLContent      string
}

type MailUser struct {
	Name  string
	Email string
}

type Options struct {
	From           MailUser
	SendGridAPIKey string
}

var MailOptions Options
var client *sendgrid.Client

func Init(options Options) {
	MailOptions = options
	client = sendgrid.NewSendClient(options.SendGridAPIKey)
}

func SendMail(params MailParams) (*rest.Response, error) {
	from := mail.NewEmail(params.From.Name, params.From.Email)
	to := mail.NewEmail(params.To.Name, params.To.Email)
	message := mail.NewSingleEmail(from, params.Subject, to, params.PlainTextContent, params.HTMLContent)
	return client.Send(message)
}
