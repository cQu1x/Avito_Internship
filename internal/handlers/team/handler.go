package team

import (
	"AvitoInternship/internal/handlers/common"
	"AvitoInternship/internal/handlers/dto"
	"database/sql"
	"encoding/json"
	"net/http"
)

func AddTeam(db *sql.DB) http.HandlerFunc {
	repo := NewTeamRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}

		var req dto.TeamDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}

		if req.TeamName == "" {
			common.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name is required")
			return
		}

		ctx := r.Context()
		team, err := repo.AddTeam(ctx, req)
		if err != nil {
			if err.Error() == dto.TeamExistsError {
				common.WriteError(w, http.StatusConflict, dto.ErrorCodeTeamExists, "team_name already exists")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to add team")
			return
		}

		response := map[string]dto.TeamDTO{"team": *team}
		common.WriteJSON(w, http.StatusOK, response)
	}
}

// GetTeam - GET /team/get?team_name=...
func GetTeam(db *sql.DB) http.HandlerFunc {
	repo := NewTeamRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}

		teamName := r.URL.Query().Get("team_name")
		if teamName == "" {
			common.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name is required")
			return
		}

		ctx := r.Context()
		team, err := repo.GetTeam(ctx, teamName)
		if err != nil {
			if err.Error() == dto.TeamNotFoundError {
				common.WriteError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "resource not found")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get team")
			return
		}

		common.WriteJSON(w, http.StatusOK, team)
	}
}
