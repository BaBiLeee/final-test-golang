package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo { return &Repo{db: db} }

// CreatePostWithLog creates post and activity_log in one transaction.
func (r *Repo) CreatePostWithLog(ctx context.Context, p *Post) (*Post, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var id int
	// Insert into posts; use pq.Array to pass string array
	err = tx.QueryRowContext(ctx,
		`INSERT INTO posts (title, content, tags) VALUES ($1, $2, $3) RETURNING id, created_at`,
		p.Title, p.Content, pq.Array(p.Tags)).Scan(&id, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	p.ID = id

	// Insert activity log
	_, err = tx.ExecContext(ctx, `INSERT INTO activity_logs (action, post_id) VALUES ($1, $2)`, "new_post", id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}

// GetPostByID
func (r *Repo) GetPostByID(ctx context.Context, id int) (*Post, error) {
	p := &Post{}
	var tags []string
	err := r.db.QueryRowContext(ctx, `SELECT id, title, content, tags, created_at FROM posts WHERE id=$1`, id).
		Scan(&p.ID, &p.Title, &p.Content, pq.Array(&tags), &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Tags = tags
	return p, nil
}

// UpdatePost
func (r *Repo) UpdatePost(ctx context.Context, id int, p *Post) error {
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET title=$1, content=$2, tags=$3 WHERE id=$4`,
		p.Title, p.Content, pq.Array(p.Tags), id)
	return err
}

// SearchByTag
func (r *Repo) SearchByTag(ctx context.Context, tag string) ([]*Post, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, content, tags, created_at FROM posts WHERE tags @> $1`,
		pq.Array([]string{tag}))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*Post
	for rows.Next() {
		p := &Post{}
		var tags []string
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, pq.Array(&tags), &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Tags = tags
		res = append(res, p)
	}
	return res, nil
}
