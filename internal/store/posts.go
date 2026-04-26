package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	UserID    int64     `json:"user_id"`
	Version   int       `json:"version"`
	UpdatedAt string    `json:"updated_at"`
	CreatedAt string    `json:"created_at"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comments_count"`
}

type PostStore struct {
	db *sql.DB
}

func (p *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, user_id, title, content, tags, updated_at, created_at, version
		FROM posts 
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	var post Post
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.UpdatedAt,
		&post.CreatedAt,
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (p *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts(title, content, tags, user_id) 
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	err := p.db.QueryRowContext(
		ctx,
		query,
		post.Title, post.Content, pq.Array(post.Tags), post.UserID,
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, tags = $3, updated_at = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	err := p.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		time.Now(),
		post.ID,
		post.Version,
	).Scan(&post.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (p *PostStore) Delete(ctx context.Context, post *Post) error {
	query := `DELETE FROM posts WHERE id = $1 AND version = $2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	res, err := p.db.ExecContext(ctx, query, post.ID, post.Version)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *PostStore) GetUserFeeds(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]*PostWithMetadata, error) {
	query := fmt.Sprintf(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at, p.version, p.tags,
			   u.id AS user_id, u.email, u.name,
			   COUNT(c.id) AS comment_count
		FROM posts p
		LEFT JOIN comments c ON p.id = c.post_id
		LEFT JOIN users u ON u.id = p.user_id
		JOIN followers f ON p.user_id = f.user_id OR p.user_id = $1
		WHERE 
		    (f.user_id = $1 OR p.user_id = $1)
		    AND (p.title ILIKE '%%' || $4 || '%%' OR p.content ILIKE '%%' || $4 || '%%')
			AND (p.tags @> $5 OR $5 = '{}')
		GROUP BY p.id, u.id, p.created_at
		ORDER BY p.created_at %s
		LIMIT $2 OFFSET $3
	`, fq.Sort)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeDuration)
	defer cancel()

	fmt.Println("fq", fq.Tags)
	rows, err := p.db.QueryContext(ctx, query, userID, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*PostWithMetadata

	for rows.Next() {
		var feed PostWithMetadata

		err = rows.Scan(
			&feed.ID,
			&feed.UserID,
			&feed.Title,
			&feed.Content,
			&feed.CreatedAt,
			&feed.UpdatedAt,
			&feed.Version,
			pq.Array(&feed.Tags),
			&feed.User.ID,
			&feed.User.Name,
			&feed.User.Email,
			&feed.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, &feed)
	}

	return feeds, nil
}
