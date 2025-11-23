package pullRequest

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"AvitoInternship/internal/handlers/dto"

	"github.com/lib/pq"
)

type PullRequestRepository struct {
	db *sql.DB
}

func NewPullRequestRepository(db *sql.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

const (
	insertPRSQL        = `INSERT INTO pull_request(id, title, author_id, status) VALUES ($1, $2, $3, $4);`
	insertReviewerSQL  = `INSERT INTO pull_request_reviewer(pr_id, user_id) VALUES ($1, $2);`
	selectUserTeamSQL  = `SELECT team_id FROM "user" WHERE id = $1;`
	selectReviewersSQL = `SELECT id FROM "user" WHERE team_id = $1 AND id != $2 AND is_active = TRUE ORDER BY random() LIMIT 2;`
	selectPRByIDSQL    = `SELECT id, title, author_id, status, created_at, updated_at FROM pull_request WHERE id = $1;`
	updatePRStatusSQL  = `UPDATE pull_request SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING id, title, author_id, status, created_at, updated_at;`
)

type PRRecord struct {
	ID          string
	Title       string
	AuthorID    string
	Status      string
	ReviewerIDs []string
}

func (pr *PRRecord) HasReviewer(id string) bool {
	for _, r := range pr.ReviewerIDs {
		if r == id {
			return true
		}
	}
	return false
}

func (r *PullRequestRepository) GetByIDForUpdateTx(ctx context.Context, tx *sql.Tx, prID string) (*PRRecord, error) {
	var rec PRRecord
	row := tx.QueryRowContext(ctx, `SELECT id, title, author_id, status FROM pull_request WHERE id = $1 FOR UPDATE`, prID)
	if err := row.Scan(&rec.ID, &rec.Title, &rec.AuthorID, &rec.Status); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM pull_request_reviewer WHERE pr_id = $1 FOR UPDATE`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		rec.ReviewerIDs = append(rec.ReviewerIDs, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *PullRequestRepository) ReassignReviewerTx(ctx context.Context, tx *sql.Tx, prID, oldReviewerID, newReviewerID string) error {
	res, err := tx.ExecContext(ctx, `DELETE FROM pull_request_reviewer WHERE pr_id = $1 AND user_id = $2`, prID, oldReviewerID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO pull_request_reviewer(pr_id, user_id) VALUES ($1, $2)`, prID, newReviewerID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pull_request SET updated_at = NOW() WHERE id = $1`, prID); err != nil {
		return err
	}
	return nil
}

func (r *PullRequestRepository) ListReviewersTx(ctx context.Context, tx *sql.Tx, prID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM pull_request_reviewer WHERE pr_id = $1 ORDER BY user_id`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reviewers []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reviewers, nil
}

func (r *PullRequestRepository) CreatePullRequest(ctx context.Context, payload dto.PullRequestDTO) (*dto.PullRequestDTO, error) {
	var teamID int
	if err := r.db.QueryRowContext(ctx, selectUserTeamSQL, payload.AuthorID).Scan(&teamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("NOT_FOUND")
		}
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, selectReviewersSQL, teamID, payload.AuthorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, insertPRSQL, payload.PullRequestID, payload.PullRequestName, payload.AuthorID, "OPEN"); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return nil, errors.New("PR_EXISTS")
		}
		return nil, err
	}

	for _, rev := range reviewers {
		if _, err := tx.ExecContext(ctx, insertReviewerSQL, payload.PullRequestID, rev); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	var id, title, authorID, status string
	var createdAt, updatedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, selectPRByIDSQL, payload.PullRequestID).Scan(&id, &title, &authorID, &status, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	res := &dto.PullRequestDTO{
		PullRequestID:     id,
		PullRequestName:   title,
		AuthorID:          authorID,
		Status:            status,
		AssignedReviewers: reviewers,
	}
	if createdAt.Valid {
		t := createdAt.Time
		res.CreatedAt = &t
	}
	return res, nil
}

func (r *PullRequestRepository) MergePullRequest(ctx context.Context, pullRequestID string) (*dto.PullRequestDTO, error) {
	var id, title, authorID, status string
	var createdAt, updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, updatePRStatusSQL, "MERGED", pullRequestID).Scan(&id, &title, &authorID, &status, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("NOT_FOUND")
		}
		return nil, err
	}

	revRows, err := r.db.QueryContext(ctx, `SELECT user_id FROM pull_request_reviewer WHERE pr_id = $1 ORDER BY user_id;`, id)
	if err != nil {
		return nil, err
	}
	defer revRows.Close()
	var reviewers []string
	for revRows.Next() {
		var uid string
		if err := revRows.Scan(&uid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, uid)
	}
	if err := revRows.Err(); err != nil {
		return nil, err
	}

	var mergedAt *time.Time
	if updatedAt.Valid {
		t := updatedAt.Time
		mergedAt = &t
	}

	res := &dto.PullRequestDTO{
		PullRequestID:     id,
		PullRequestName:   title,
		AuthorID:          authorID,
		Status:            status,
		AssignedReviewers: reviewers,
		MergedAt:          mergedAt,
	}
	return res, nil
}
