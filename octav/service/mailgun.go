package service

import (
	"os"
	"strings"
	"sync"

	"github.com/builderscon/octav/octav/log"
	pdebug "github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

var mailgunSvc MailgunSvc
var mailgunOnce sync.Once

type MailLog struct {
	MessageID      string   `json:"message_id"`
	ServerResponse string   `json:"server_response"`
	From           string   `json:"from"`
	Recipients     []string `json:"recipients"`
	Subject        string   `json:"subject"`
	Text           string   `json:"text"`
}

func Mailgun() *MailgunSvc {
	mailgunOnce.Do(mailgunSvc.Init)
	return &mailgunSvc
}

func (v *MailgunSvc) Init() {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Mailgun.Init")
		defer g.End()
	}

	f := func(v *string, envname string) {
		envvar := os.Getenv(envname)
		if envvar == "" {
			panic("Missing required environment variable " + envname)
		}
		*v = envvar
	}

	f(&v.defaultSender, "MAILGUN_DEFAULT_SENDER")

	var domain string
	var apiKey string
	var publicApiKey string
	f(&domain, "MAILGUN_DOMAIN")
	f(&apiKey, "MAILGUN_SECRET_API_KEY")
	f(&publicApiKey, "MAILGUN_PUBLIC_API_KEY")

	if pdebug.Enabled {
		pdebug.Printf(
			"Creating Mailgun client with domain=%s, apiKey=%s, publicApiKey=%s",
			domain,
			strings.Repeat("*", len(apiKey)-4)+apiKey[len(apiKey)-4:],
			strings.Repeat("*", len(publicApiKey)-4)+publicApiKey[len(publicApiKey)-4:],
		)
	}

	v.client = mailgun.NewMailgun(domain, apiKey, publicApiKey)
}

type MailMessage struct {
	From       string
	Subject    string
	Text       string
	Recipients []string
}

func (v *MailgunSvc) Send(mm *MailMessage) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Mailgun.Send").BindError(&err)
		defer g.End()
	}

	if mm.From == "" {
		mm.From = v.defaultSender
	}

	m := mailgun.NewMessage(mm.From, mm.Subject, mm.Text, mm.Recipients...)

	mg := v.client
	msg, id, err := mg.Send(m)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	log.Notice(interface{}(&MailLog{
		MessageID:      id,
		ServerResponse: msg,
		From:           mm.From,
		Recipients:     mm.Recipients,
		Subject:        mm.Subject,
		Text:           mm.Text,
	}))
	return nil
}
