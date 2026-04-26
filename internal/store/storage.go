package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("record not found")
var QueryTimeDuration = 5 * time.Second

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, *Post) error
		Update(context.Context, *Post) error
		GetUserFeeds(context.Context, int64, PaginatedFeedQuery) ([]*PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context, *sql.Tx, *User) error
		GetByID(context.Context, int64) (*User, error)
		GetByEmail(context.Context, string) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		CreateUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error
		Activate(context.Context, string) error
		Delete(context.Context, *User) error
	}
	Comment interface {
		GetByPostID(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}
	Followers interface {
		Follow(ctx context.Context, followerID, userID int64) error
		Unfollow(ctx context.Context, followerID, userID int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db: db},
		Users:     &UserStore{db: db},
		Comment:   &CommentStore{db: db},
		Followers: &FollowerStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, f func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
