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

var testMailer = mailer.New(mailer.Config{
	Domain:      "localhost",
	Templates:   "./testdata/mail",
	FromAddress: "me@here.com",
	FromName:    "Joe",
	JobsSize:    1,
	ResultsSize: 1,
})

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

	testMailer.Host = host
	testMailer.Port = port.Int()

	time.Sleep(2 * time.Second)

	go testMailer.ListenForMail()

	code := m.Run()

	if err := testcontainers.TerminateContainer(mailhog); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}
