package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

// Message is the type for an email message
type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

// Result contains information regarding the status of the sent email message
type Result struct {
	Success bool
	Error   error
}

type Config struct {
	TemplatesFS fs.FS
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	BaseURL     string
	JobsSize    int
	ResultsSize int
}

// Mailer holds the information necessary to connect to an SMTP server
type Mailer struct {
	TemplatesFS fs.FS
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	BaseURL     string
	Jobs        chan Message
	Results     chan Result
}

func New(cfg Config) *Mailer {
	return &Mailer{
		TemplatesFS: cfg.TemplatesFS,
		Host:        cfg.Host,
		Port:        cfg.Port,
		Username:    cfg.Username,
		Password:    cfg.Password,
		Encryption:  cfg.Encryption,
		FromAddress: cfg.FromAddress,
		FromName:    cfg.FromName,
		BaseURL:     cfg.BaseURL,
		Jobs:        make(chan Message, cfg.JobsSize),
		Results:     make(chan Result, cfg.ResultsSize),
	}
}

// ListenForMail listens to the mail channel and sends mail
// when it receives a payload. It runs continually in the background,
// and sends error/success messages back on the Results channel.
func (m *Mailer) ListenForMail() {
	for {
		msg := <-m.Jobs
		err := m.SendSMTPMessage(msg)
		if err != nil {
			m.Results <- Result{false, err}
		} else {
			m.Results <- Result{true, nil}
		}
	}
}

func (m *Mailer) SendSMTPMessage(msg Message) error {
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

// getEncryption returns the appropriate encryption type based on a string value
func (m *Mailer) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	default:
		return mail.EncryptionNone
	}
}

// buildHTMLMessage creates the html version of the message
func (m *Mailer) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s.html.tmpl", msg.Template)
	t, err := template.New("email-html").ParseFS(m.TemplatesFS, templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

// buildPlainTextMessage creates the plaintext version of the message
func (m *Mailer) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s.plain.tmpl", msg.Template)
	t, err := template.New("email-html").ParseFS(m.TemplatesFS, templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

// inlineCSS takes html input as a string, and inlines css where possible
func (m *Mailer) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
