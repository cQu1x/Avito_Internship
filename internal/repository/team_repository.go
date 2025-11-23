package repository

import "database/sql"

type TeamRepository struct {
	db *sql.DB
}

// NewTeamRepository создает новый TeamRepository
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}
