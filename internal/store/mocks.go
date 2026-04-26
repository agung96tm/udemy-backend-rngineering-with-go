package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
	users []User
}

func (m *MockUserStore) Create(context.Context, *sql.Tx, *User) error {
	return nil
}

func (m *MockUserStore) GetByID(context.Context, int64) (*User, error) {
	return &User{
		ID: 99,
	}, nil
}

func (m *MockUserStore) GetByEmail(context.Context, string) (*User, error) {
	return nil, nil
}
func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}
func (m *MockUserStore) CreateUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	return nil
}
func (m *MockUserStore) Activate(context.Context, string) error {
	return nil
}
func (m *MockUserStore) Delete(context.Context, *User) error {
	return nil
}
