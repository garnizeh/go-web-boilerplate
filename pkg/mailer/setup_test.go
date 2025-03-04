package mailer_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/garnizeh/go-web-boilerplate/pkg/mailer"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var mail = mailer.Mail{
	Domain:      "localhost",
	Templates:   "./testdata/mail",
	Encryption:  "none",
	FromAddress: "me@here.com",
	FromName:    "Joe",
	Jobs:        make(chan mailer.Message, 1),
	Results:     make(chan mailer.Result, 1),
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	mailhog, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mailhog/mailhog",
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor:   wait.ForListeningPort("1025/tcp"),
			Name:         "mailhog",
		},
		Started: true,
	})
	if err != nil {
		log.Print(err)
		return
	}

    host, err := mailhog.Host(ctx)
    if err != nil {
        log.Print(err)
        return
    }

    port, err := mailhog.MappedPort(ctx, "1025")
    if err != nil {
        log.Print(err)
        return
    }

    mail.Host = host
	mail.Port = port.Int()

	time.Sleep(2 * time.Second)

	go mail.ListenForMail()

	code := m.Run()

	if err := testcontainers.TerminateContainer(mailhog); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}
