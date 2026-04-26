package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Follower struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followerID int64, userID int64) error {
	if followerID == userID {
		return errors.New("cannot follow yourself")
	}

	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, follower_id) DO NOTHING;
	`

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}
	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerID int64, userID int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2;
	`

	res, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return nil
	}
	return nil
}
