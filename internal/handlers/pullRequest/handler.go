package pullRequest

import (
	"AvitoInternship/internal/handlers/common"
	"AvitoInternship/internal/handlers/dto"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func Create(db *sql.DB) http.HandlerFunc {
	repo := NewPullRequestRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteError(w, http.StatusMethodNotAllowed, dto.ErrorMethodNotAllowed, "method not allowed")
			return
		}
		var req dto.PullRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, dto.ErrorBadRequest, "invalid request body")
			return
		}
		if req.PullRequestName == "" || req.AuthorID == "" {
			common.WriteError(w, http.StatusBadRequest, dto.ErrorBadRequest, "pull_request_name and author_id are required")
			return
		}
		ctx := r.Context()
		pr, err := repo.CreatePullRequest(ctx, req)
		if err != nil {
			if err.Error() == dto.ErrorCodeNotFound {
				common.WriteError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "resource not found")
				return
			}
			if err.Error() == dto.ErrorCodePRExists {
				common.WriteError(w, http.StatusConflict, dto.ErrorCodePRExists, "PR id already exists")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, dto.ErrorInternalError, "failed to create pull request")
			return
		}

		common.WriteJSON(w, http.StatusCreated, map[string]*dto.PullRequestDTO{"pr": pr})
	}
}

func Merge(db *sql.DB) http.HandlerFunc {
	repo := NewPullRequestRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteError(w, http.StatusMethodNotAllowed, dto.ErrorMethodNotAllowed, "method not allowed")
			return
		}
		var req dto.MergePRRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, dto.ErrorBadRequest, "invalid request body")
			return
		}
		if req.PullRequestID == "" {
			common.WriteError(w, http.StatusBadRequest, dto.ErrorBadRequest, "pull_request_id is required")
			return
		}
		ctx := r.Context()
		pr, err := repo.MergePullRequest(ctx, req.PullRequestID)
		if err != nil {
			if err.Error() == dto.ErrorCodeNotFound {
				common.WriteError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "resource not found")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, dto.ErrorInternalError, "failed to merge pull request")
			return
		}
		common.WriteJSON(w, http.StatusOK, map[string]*dto.PullRequestDTO{"pr": pr})
	}
}

func Reassign(db *sql.DB) http.HandlerFunc {
	prRepo := NewPullRequestRepository(db)
	userRepository := &userRepository{db: db}
	svc := NewPullRequestService(prRepo, userRepository)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteError(w, http.StatusMethodNotAllowed, dto.ErrorMethodNotAllowed, "method not allowed")
			return
		}
		var req dto.ReassignReviewerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, dto.ErrorBadRequest, "invalid request body")
			return
		}
		if req.PullRequestID == "" || req.OldReviewerID == "" {
			common.WriteError(w, http.StatusBadRequest, dto.ErrorBadRequest, "pull_request_id and old_reviewer_id are required")
			return
		}
		ctx := r.Context()
		res, err := svc.Reassign(ctx, req.PullRequestID, req.OldReviewerID)
		if err != nil {
			if err.Error() == dto.ErrorCodeNotFound {
				common.WriteError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "resource not found")
				return
			}
			if err.Error() == dto.ErrorCodePRMerged {
				common.WriteError(w, http.StatusBadRequest, dto.ErrorCodePRMerged, "cannot reassign on merged PR")
				return
			}
			common.WriteError(w, http.StatusBadRequest, dto.ErrorInternalError, "failed to reassign reviewer")
			return
		}

		out := map[string]interface{}{
			"pr":          res.PR,
			"replaced_by": res.ReplacedBy,
		}
		common.WriteJSON(w, http.StatusOK, out)
	}
}

type userRepository struct {
	db *sql.DB
}

func (s *userRepository) GetByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	var id, name string
	var teamID int
	var isActive bool
	row := s.db.QueryRowContext(ctx, `SELECT id, name, team_id, is_active FROM "user" WHERE id = $1`, userID)
	if err := row.Scan(&id, &name, &teamID, &isActive); err != nil {
		return nil, err
	}
	return &dto.UserDTO{UserID: id, Username: name, TeamID: teamID, IsActive: isActive}, nil
}

func (s *userRepository) GetRandomActiveTeammate(ctx context.Context, teamID int, exclude []string) (*dto.UserDTO, error) {
	base := `SELECT id, name, team_id, is_active FROM "user" WHERE team_id = $1 AND is_active = TRUE`
	args := []interface{}{teamID}
	if len(exclude) > 0 {
		var placeholders []string
		for i := range exclude {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
			args = append(args, exclude[i])
		}
		base = base + " AND id NOT IN (" + strings.Join(placeholders, ",") + ")"
	}
	base = base + " ORDER BY random() LIMIT 1"
	row := s.db.QueryRowContext(ctx, base, args...)
	var id, name string
	var tid int
	var isActive bool
	if err := row.Scan(&id, &name, &tid, &isActive); err != nil {
		return nil, err
	}
	return &dto.UserDTO{UserID: id, Username: name, TeamID: tid, IsActive: isActive}, nil
}
