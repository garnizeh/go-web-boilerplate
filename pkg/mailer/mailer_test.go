package mailer_test

import (
	"errors"
	"testing"

	"github.com/garnizeh/go-web-boilerplate/pkg/mailer"
)

func TestMailSendSMTPMessage(t *testing.T) {
	msg := mailer.Message{
		From:        "me@here.com",
		FromName:    "Joe",
		To:          "you@there.com",
		Subject:     "test",
		Template:    "test",
		Attachments: []string{"testdata/test.html.tmpl"},
	}

	err := testMailer.SendSMTPMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMailSendUsingChan(t *testing.T) {
	msg := mailer.Message{
		From:        "me@here.com",
		FromName:    "Joe",
		To:          "you@there.com",
		Subject:     "test",
		Template:    "test",
		Attachments: []string{"testdata/test.html.tmpl"},
	}

	testMailer.Jobs <- msg
	res := <-testMailer.Results
	if res.Error != nil {
		t.Error(errors.New("failed to send over channel"))
	}

	msg.To = "not_an_email_address"
	testMailer.Jobs <- msg
	res = <-testMailer.Results
	if res.Error == nil {
		t.Error(errors.New("no error received with invalid to address"))
	}
}
