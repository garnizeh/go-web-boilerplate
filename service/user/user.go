package user

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/garnizeh/go-web-boilerplate/pkg/mailer"
	"github.com/garnizeh/go-web-boilerplate/pkg/securepass"
	"github.com/garnizeh/go-web-boilerplate/storage"
	"github.com/garnizeh/go-web-boilerplate/storage/datastore"
	"github.com/google/uuid"
)

const (
	tokenSignup   = "SIGNUP"
	tokenPassword = "PASSWORD"

	tokenDurationSignup   = time.Hour * 12
	tokenDurationPassword = time.Hour * 1
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailInUse          = errors.New("email already in use")
	ErrUserNotVerified     = errors.New("email not verified")
	ErrEmailNotFound       = errors.New("email not found")
	ErrUserAlreadyVerified = errors.New("user already verified")
	ErrInvalidToken        = errors.New("invalid token")
)

type Service struct {
	securepass *securepass.Securepass
	mailer     *mailer.Mailer
	db         *storage.DB[datastore.Queries]
	userCache  *sync.Map
}

func New(
	securepass *securepass.Securepass,
	mailer *mailer.Mailer,
	db *storage.DB[datastore.Queries],
) *Service {
	return &Service{
		securepass: securepass,
		mailer:     mailer,
		db:         db,
		userCache:  &sync.Map{},
	}
}

type User struct {
	Name  string
	Email string
}

func (s *Service) GetUser(ctx context.Context, email string) (User, error) {
	if v, ok := s.userCache.Load(email); ok {
		return v.(User), nil
	}

	var user datastore.User
	if err := s.db.Read(ctx, func(queries *datastore.Queries) error {
		var err error
		user, err = queries.GetUser(ctx, email)
		return err
	}); err != nil {
		return User{}, err
	}

	return s.updateCache(user), nil
}

func (s *Service) Signin(
	ctx context.Context,
	email string,
	password string,
) (User, error) {
	var user datastore.User
	if err := s.db.Read(ctx, func(queries *datastore.Queries) error {
		var err error
		user, err = queries.GetUser(ctx, email)
		if err != nil {
			if storage.NoRows(err) {
				return ErrInvalidCredentials
			}

			return err
		}

		if user.VerifiedAt == 0 {
			return ErrUserNotVerified
		}

		return nil
	}); err != nil {
		return User{}, err
	}

	if err := s.securepass.Compare(user.Password, user.Salt, []byte(password)); err != nil {
		return User{}, ErrInvalidCredentials
	}

	return s.updateCache(user), nil
}

func (s *Service) Signup(
	ctx context.Context,
	baseURL string,
	email string,
	name string,
	password string,
) error {
	if err := s.db.Write(ctx, func(queries *datastore.Queries) error {
		_, err := queries.GetUser(ctx, email)
		if err == nil {
			return ErrEmailInUse
		}
		if !storage.NoRows(err) {
			return fmt.Errorf("failed to check for the email existence in the database: %w", err)
		}

		hashSalt, err := s.securepass.GenerateHash([]byte(password), nil)
		if err != nil {
			return fmt.Errorf("failed to hash the password: %w", err)
		}

		if err := queries.CreateUser(ctx, datastore.CreateUserParams{
			Email:    email,
			Name:     name,
			Password: hashSalt.Hash,
			Salt:     hashSalt.Salt,
		}); err != nil {
			return fmt.Errorf("failed to create the user in the database: %w", err)
		}

		if err := queries.DeleteSignupTokensByEmail(ctx, email); err != nil {
			return fmt.Errorf("failed to delete existing signup tokens for the email %q in the database: %w", email, err)
		}

		token := uuid.New().String()
		expiresAt := time.Now().Add(tokenDurationSignup).UTC().UnixMilli()
		if err := queries.CreateToken(ctx, datastore.CreateTokenParams{
			Token:     token,
			Type:      tokenSignup,
			Email:     email,
			ExpiresAt: expiresAt,
		}); err != nil {
			return fmt.Errorf("failed to create the signup token in the database: %w", err)
		}

		// mail := mailer.NewMailSignup(baseURL, email, name, token)
		// if err := s.mailer.SendMailSignup(mail); err != nil {
		// 	return fmt.Errorf("failed to send the signup confirmation email: %w", err)
		// }

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) ResendSignupToken(
	ctx context.Context,
	baseURL string,
	email string,
) error {
	if err := s.db.Write(ctx, func(queries *datastore.Queries) error {
		user, err := queries.GetUser(ctx, email)
		if err != nil {
			if storage.NoRows(err) {
				return ErrUserNotVerified
			}

			return fmt.Errorf("failed to check for the email existence in the database: %w", err)
		}
		if user.VerifiedAt > 0 {
			return ErrUserAlreadyVerified
		}

		if err := queries.DeleteSignupTokensByEmail(ctx, email); err != nil {
			return fmt.Errorf("failed to delete existing signup tokens for the email %q in the database: %w", email, err)
		}

		token := uuid.New().String()
		expiresAt := time.Now().Add(tokenDurationSignup).UTC().UnixMilli()
		if err := queries.CreateToken(ctx, datastore.CreateTokenParams{
			Token:     token,
			Type:      tokenSignup,
			Email:     email,
			ExpiresAt: expiresAt,
		}); err != nil {
			return fmt.Errorf("failed to create the signup token in the database: %w", err)
		}

		// mail := mailer.NewMailSignup(baseURL, email, user.Name, token)
		// if err := s.mailer.SendMailSignup(mail); err != nil {
		// 	return fmt.Errorf("failed to send the signup confirmation email: %w", err)
		// }

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) ValidateSignupToken(
	ctx context.Context,
	token string,
) (datastore.User, error) {
	var user datastore.User
	if err := s.db.Read(ctx, func(queries *datastore.Queries) error {
		now := time.Now().UTC().UnixMilli()
		registeredToken, err := queries.GetSignupTokenNotExpired(ctx, datastore.GetSignupTokenNotExpiredParams{
			Token:     token,
			ExpiresAt: now,
		})
		if err != nil {
			if storage.NoRows(err) {
				return ErrInvalidToken
			}

			return err
		}

		user, err = queries.SetUserIsVerified(ctx, registeredToken.Email)
		if err != nil {
			return err
		}

		if err := queries.DeleteSignupTokensByEmail(ctx, registeredToken.Email); err != nil {
			return err
		}

		_ = s.updateCache(user)

		return nil
	}); err != nil {
		return datastore.User{}, err
	}

	return user, nil
}

func (s *Service) ResetPassword(
	ctx context.Context,
	baseURL string,
	email string,
) error {
	if err := s.db.Write(ctx, func(queries *datastore.Queries) error {
		user, err := queries.GetUser(ctx, email)
		if err != nil {
			if storage.NoRows(err) {
				return ErrEmailNotFound
			}

			return fmt.Errorf("failed to check for the email existence in the database: %w", err)
		}
		if user.VerifiedAt == 0 {
			return ErrUserNotVerified
		}

		if err := queries.DeletePasswordTokensByEmail(ctx, email); err != nil {
			return fmt.Errorf("failed to delete existing reset password tokens for the email %q in the database: %w", email, err)
		}

		token := uuid.New().String()
		expiresAt := time.Now().Add(tokenDurationPassword).UTC().UnixMilli()
		if err := queries.CreateToken(ctx, datastore.CreateTokenParams{
			Token:     token,
			Type:      tokenPassword,
			Email:     email,
			ExpiresAt: expiresAt,
		}); err != nil {
			return fmt.Errorf("failed to create the reset password token in the database: %w", err)
		}

		// mail := mailer.NewMailPassword(baseURL, email, user.Name, token)
		// if err := s.mailer.SendMailSignup(mail); err != nil {
		// 	return fmt.Errorf("failed to send the reset password email: %w", err)
		// }

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) ResetPasswordToken(
	ctx context.Context,
	token string,
) error {
	if err := s.db.Read(ctx, func(queries *datastore.Queries) error {
		now := time.Now().UTC().UnixMilli()
		registeredToken, err := queries.GetPasswordTokenNotExpired(ctx, datastore.GetPasswordTokenNotExpiredParams{
			Token:     token,
			ExpiresAt: now,
		})
		if err != nil {
			if storage.NoRows(err) {
				return ErrInvalidToken
			}

			return err
		}

		if _, err = queries.GetUser(ctx, registeredToken.Email); err != nil {
			if storage.NoRows(err) {
				return ErrEmailNotFound
			}

			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) ChangePassword(
	ctx context.Context,
	email string,
	currentPassword string,
	newPassword string,
) error {
	var user datastore.User
	if err := s.db.Read(ctx, func(queries *datastore.Queries) error {
		var err error
		user, err = queries.GetUser(ctx, email)
		if err != nil {
			if storage.NoRows(err) {
				return ErrInvalidCredentials
			}

			return err
		}

		if user.VerifiedAt == 0 {
			return ErrUserNotVerified
		}

		if err := s.securepass.Compare(user.Password, user.Salt, []byte(currentPassword)); err != nil {
			return ErrInvalidCredentials
		}

		hashSalt, err := s.securepass.GenerateHash([]byte(newPassword), nil)
		if err != nil {
			return fmt.Errorf("failed to hash the password: %w", err)
		}

		user, err = queries.UpdateUserPassword(ctx, datastore.UpdateUserPasswordParams{
			Email:    email,
			Password: hashSalt.Hash,
			Salt:     hashSalt.Salt,
		})
		if err != nil {
			if storage.NoRows(err) {
				return ErrEmailNotFound
			}

			return err
		}

		_ = s.updateCache(user)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) ChangePasswordWithToken(
	ctx context.Context,
	token string,
	password string,
) (datastore.User, error) {
	var user datastore.User
	if err := s.db.Read(ctx, func(queries *datastore.Queries) error {
		now := time.Now().UTC().UnixMilli()
		registeredToken, err := queries.GetPasswordTokenNotExpired(ctx, datastore.GetPasswordTokenNotExpiredParams{
			Token:     token,
			ExpiresAt: now,
		})
		if err != nil {
			if storage.NoRows(err) {
				return ErrInvalidToken
			}

			return err
		}

		hashSalt, err := s.securepass.GenerateHash([]byte(password), nil)
		if err != nil {
			return fmt.Errorf("failed to hash the password: %w", err)
		}

		user, err = queries.UpdateUserPassword(ctx, datastore.UpdateUserPasswordParams{
			Email:    registeredToken.Email,
			Password: hashSalt.Hash,
			Salt:     hashSalt.Salt,
		})
		if err != nil {
			if storage.NoRows(err) {
				return ErrEmailNotFound
			}

			return err
		}

		if err := queries.DeletePasswordTokensByEmail(ctx, registeredToken.Email); err != nil {
			return err
		}

		_ = s.updateCache(user)

		return nil
	}); err != nil {
		return datastore.User{}, err
	}

	return user, nil
}

func (s *Service) updateCache(u datastore.User) User {
	user := User{
		Name:  u.Name,
		Email: u.Email,
	}
	s.userCache.LoadOrStore(u.Email, user)

	return user
}
