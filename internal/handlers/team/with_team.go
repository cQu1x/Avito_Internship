package team

import (
	"AvitoInternship/internal/handlers/dto"
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

const insertTeamSQL = `
INSERT INTO team(name) VALUES ($1) RETURNING id;
`

const insertUserSQL = `
INSERT INTO "user"(id, name, is_active, team_id) VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, is_active = EXCLUDED.is_active, team_id = EXCLUDED.team_id;
`

const selectTeamIDSQL = `
SELECT id FROM team WHERE name = $1;
`

const selectUsersByTeamSQL = `
SELECT id, name, is_active FROM "user" WHERE team_id = $1 ORDER BY id;
`

func (r *TeamRepository) AddTeam(ctx context.Context, t dto.TeamDTO) (*dto.TeamDTO, error) {
	transaction, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = transaction.Rollback() }()

	var teamID int
	err = transaction.QueryRowContext(ctx, insertTeamSQL, t.TeamName).Scan(&teamID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return nil, errors.New(dto.TeamExistsError)
		}
		return nil, err
	}

	for _, m := range t.Members {
		if _, err := transaction.ExecContext(ctx, insertUserSQL, m.UserID, m.Username, m.IsActive, teamID); err != nil {
			return nil, err
		}
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
	var teamID int
	err := r.db.QueryRowContext(ctx, selectTeamIDSQL, teamName).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(dto.TeamNotFoundError)
		}
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, selectUsersByTeamSQL, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]dto.TeamMemberDTO, 0)
	for rows.Next() {
		var m dto.TeamMemberDTO
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &dto.TeamDTO{TeamName: teamName, Members: members}, nil
}
