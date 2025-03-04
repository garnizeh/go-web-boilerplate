package service

import (
	"errors"
	"strings"

	"github.com/garnizeh/go-web-boilerplate/pkg/mailer"
	"github.com/garnizeh/go-web-boilerplate/pkg/securepass"
	"github.com/garnizeh/go-web-boilerplate/service/user"
	"github.com/garnizeh/go-web-boilerplate/storage"
	"github.com/garnizeh/go-web-boilerplate/storage/datastore"
)

type Service struct {
	user *user.Service
}

func New(
	argon *securepass.Securepass,
	mailer *mailer.Mailer,
	db *storage.DB[datastore.Queries],
) *Service {
	user := user.New(argon, mailer, db)

	return &Service{
		user: user,
	}
}

func (s *Service) User() *user.Service {
	return s.user
}

var (
	ErrInvalidParam = errors.New("invalid param")
	ErrUniqueParam  = errors.New("param violated unique constraint")
	ErrNotFound     = errors.New("found no record")
)

func CheckErr(err error) error {
	if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		return errors.Join(err, ErrUniqueParam)
	}

	return err
}
