package user

import (
	"AvitoInternship/internal/handlers/dto"

	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const updateUserIsActiveSQL = `
UPDATE "user"
SET is_active = $1
WHERE id = $2;
`

const selectUserWithTeamSQL = `
SELECT
    u.id,
    u.name,
    t.name AS team_name,
    u.is_active
FROM "user" u
JOIN team t ON t.id = u.team_id
WHERE u.id = $1;
`
const getReviewSQL = `
SELECT
    pr.id,
    pr.title,
    pr.author_id,
    pr.status
FROM pull_request pr
JOIN pull_request_reviewer prr ON pr.id = prr.pr_id
WHERE prr.user_id = $1
ORDER BY pr.id;
`

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*dto.UserDTO, error) {
	transaction, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = transaction.Rollback()
	}()

	res, err := transaction.ExecContext(ctx, updateUserIsActiveSQL, isActive, userID)
	if err != nil {
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, errors.New(userNotFoundError)
	}
	var user dto.UserDTO
	err = transaction.QueryRowContext(ctx, selectUserWithTeamSQL, userID).
		Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetReviewPullRequests(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error) {
	rows, err := r.db.QueryContext(ctx, getReviewSQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.PullRequestShortDTO

	for rows.Next() {
		var pr dto.PullRequestShortDTO
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
		); err != nil {
			return nil, err
		}
		result = append(result, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	var id, name string
	var teamID int
	var isActive bool
	row := r.db.QueryRowContext(ctx, `SELECT id, name, team_id, is_active FROM "user" WHERE id = $1`, userID)
	if err := row.Scan(&id, &name, &teamID, &isActive); err != nil {
		return nil, err
	}
	return &dto.UserDTO{UserID: id, Username: name, TeamName: "", IsActive: isActive}, nil
}

func (r *UserRepository) GetRandomActiveTeammate(ctx context.Context, teamID int, exclude []string) (*dto.UserDTO, error) {
	base := `SELECT id, name, team_id FROM "user" WHERE team_id = $1 AND is_active = TRUE`
	args := []interface{}{teamID}
	if len(exclude) > 0 {
		var placeholders []string
		for i, ex := range exclude {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
			args = append(args, ex)
		}
		base = base + " AND id NOT IN (" + strings.Join(placeholders, ",") + ")"
	}
	base = base + " ORDER BY random() LIMIT 1"

	row := r.db.QueryRowContext(ctx, base, args...)
	var id, name string
	var tid int
	if err := row.Scan(&id, &name, &tid); err != nil {
		return nil, err
	}
	return &dto.UserDTO{UserID: id, Username: name, TeamName: "", IsActive: true}, nil
}
