package mailer_test

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"testing"
	"time"

	"github.com/garnizeh/go-web-boilerplate/pkg/mailer"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed testdata/*
var testdata embed.FS
var testMailer *mailer.Mailer

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

	templateFS, err := fs.Sub(testdata, "testdata")
	if err != nil {
		log.Print(err)
		return
	}

	testMailer = mailer.New(mailer.Config{
		TemplatesFS: templateFS,
		Host:        host,
		Port:        port.Int(),
		FromAddress: "me@here.com",
		FromName:    "Joe",
		JobsSize:    1,
		ResultsSize: 1,
	})

	time.Sleep(2 * time.Second)

	go testMailer.ListenForMail()

	code := m.Run()

	if err := testcontainers.TerminateContainer(mailhog); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}
