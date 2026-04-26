package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
	UserID  int64  `json:"user_id"`
	PostID  int64  `json:"post_id"`

	User User `json:"user"`

	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	query := `
		SELECT comments.id, comments.content, comments.user_id, post_id, comments.updated_at, comments.created_at,
		users.id, users.name, users.email
		FROM comments
		JOIN users ON users.id = comments.user_id
		WHERE comments.post_id = $1
		ORDER BY comments.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		comment.User = User{}
		err = rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.UserID,
			&comment.PostID,
			&comment.UpdatedAt,
			&comment.CreatedAt,
			&comment.User.ID,
			&comment.User.Name,
			&comment.User.Email,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments(content, user_id, post_id) 
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.Content,
		comment.UserID,
		comment.PostID,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
