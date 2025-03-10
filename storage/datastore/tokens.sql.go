// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: tokens.sql

package datastore

import (
	"context"
)

const createToken = `-- name: CreateToken :exec
INSERT INTO tokens (token, type, email, expires_at)
            VALUES (?    , ?   , ?    , ?)
`

type CreateTokenParams struct {
	Token     string
	Type      string
	Email     string
	ExpiresAt int64
}

func (q *Queries) CreateToken(ctx context.Context, arg CreateTokenParams) error {
	_, err := q.db.ExecContext(ctx, createToken,
		arg.Token,
		arg.Type,
		arg.Email,
		arg.ExpiresAt,
	)
	return err
}

const deleteExpiredTokens = `-- name: DeleteExpiredTokens :exec
UPDATE tokens SET deleted_at = CAST(unixepoch('subsecond') * 1000 AS INTEGER)
WHERE expires_at <= ?
`

func (q *Queries) DeleteExpiredTokens(ctx context.Context, expiresAt int64) error {
	_, err := q.db.ExecContext(ctx, deleteExpiredTokens, expiresAt)
	return err
}

const deletePasswordTokensByEmail = `-- name: DeletePasswordTokensByEmail :exec
UPDATE tokens SET deleted_at = CAST(unixepoch('subsecond') * 1000 AS INTEGER)
WHERE email = ? AND type = 'PASSWORD'
`

func (q *Queries) DeletePasswordTokensByEmail(ctx context.Context, email string) error {
	_, err := q.db.ExecContext(ctx, deletePasswordTokensByEmail, email)
	return err
}

const deleteSignupTokensByEmail = `-- name: DeleteSignupTokensByEmail :exec
UPDATE tokens SET deleted_at = CAST(unixepoch('subsecond') * 1000 AS INTEGER)
WHERE email = ? AND type = 'SIGNUP'
`

func (q *Queries) DeleteSignupTokensByEmail(ctx context.Context, email string) error {
	_, err := q.db.ExecContext(ctx, deleteSignupTokensByEmail, email)
	return err
}

const getPasswordTokenNotExpired = `-- name: GetPasswordTokenNotExpired :one
SELECT token, type, email, expires_at, deleted_at FROM tokens
WHERE token = ? AND type = 'PASSWORD' AND expires_at >= ? AND deleted_at = 0
`

type GetPasswordTokenNotExpiredParams struct {
	Token     string
	ExpiresAt int64
}

func (q *Queries) GetPasswordTokenNotExpired(ctx context.Context, arg GetPasswordTokenNotExpiredParams) (Token, error) {
	row := q.db.QueryRowContext(ctx, getPasswordTokenNotExpired, arg.Token, arg.ExpiresAt)
	var i Token
	err := row.Scan(
		&i.Token,
		&i.Type,
		&i.Email,
		&i.ExpiresAt,
		&i.DeletedAt,
	)
	return i, err
}

const getSignupTokenNotExpired = `-- name: GetSignupTokenNotExpired :one
SELECT token, type, email, expires_at, deleted_at FROM tokens
WHERE token = ? AND type = 'SIGNUP' AND expires_at >= ? AND deleted_at = 0
`

type GetSignupTokenNotExpiredParams struct {
	Token     string
	ExpiresAt int64
}

func (q *Queries) GetSignupTokenNotExpired(ctx context.Context, arg GetSignupTokenNotExpiredParams) (Token, error) {
	row := q.db.QueryRowContext(ctx, getSignupTokenNotExpired, arg.Token, arg.ExpiresAt)
	var i Token
	err := row.Scan(
		&i.Token,
		&i.Type,
		&i.Email,
		&i.ExpiresAt,
		&i.DeletedAt,
	)
	return i, err
}
