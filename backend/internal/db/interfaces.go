package db

import (
	"context"
)

type UserQuerier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (User, error)
}
