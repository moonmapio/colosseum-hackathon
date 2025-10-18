package messages

import (
	"bytes"
	"text/template"

	"github.com/mailjet/mailjet-apiv3-go/v4"
	"moonmap.io/go-commons/helpers"
)

type MailjetManager struct {
	client    *mailjet.Client
	tmpl      *template.Template
	fromEmail string
	fromName  string
}

type EmailRecipient struct {
	Email string
	Name  string
}

func New(fromEmail, fromName string) *MailjetManager {
	pub := helpers.GetEnvOrFail("MESSAGE_MJ_APIKEY_PUBLIC")
	priv := helpers.GetEnvOrFail("MESSAGE_MJ_APIKEY_PRIVATE")

	m := &MailjetManager{
		client:    mailjet.NewMailjetClient(pub, priv),
		fromEmail: fromEmail,
		fromName:  fromName,
	}

	return m
}

func (m *MailjetManager) SetTemplate(name string, tmplt *string) error {
	tmpl, err := template.New(name).Parse(*tmplt)
	if err != nil {
		return err
	}
	m.tmpl = tmpl
	return nil
}

func (m *MailjetManager) SendMail(recipients []*EmailRecipient, subject, textBody string, data any) error {

	var buf bytes.Buffer
	if err := m.tmpl.Execute(&buf, data); err != nil {
		return err
	}
	htmlBody := buf.String()
	messagesInfo := []mailjet.InfoMessagesV31{}

	for _, rec := range recipients {
		info := mailjet.InfoMessagesV31{
			From:     &mailjet.RecipientV31{Email: m.fromEmail, Name: m.fromName},
			To:       &mailjet.RecipientsV31{{Email: rec.Email, Name: rec.Name}},
			Subject:  subject,
			TextPart: textBody,
			HTMLPart: htmlBody,
		}
		messagesInfo = append(messagesInfo, info)
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := m.client.SendMailV31(&messages)
	return err
}
