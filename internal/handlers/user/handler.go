package user

import (
	"AvitoInternship/internal/handlers/common"
	"AvitoInternship/internal/handlers/dto"
	"database/sql"
	"encoding/json"
	"net/http"
)

const (
	ErrorCodeNotFound = "NOT_FOUND"
	userNotFoundError = "USER_NOT_FOUND"
)

func SetIsActive(db *sql.DB) http.HandlerFunc {
	repo := NewUserRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			common.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}

		var req dto.SetIsActiveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}

		if req.UserID == "" {
			common.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
			return
		}

		ctx := r.Context()
		userDTO, err := repo.SetIsActive(ctx, req.UserID, req.IsActive)
		if err != nil {
			if err.Error() == userNotFoundError {
				common.WriteError(w, http.StatusNotFound, ErrorCodeNotFound, "resource not found")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user")
			return
		}

		response := dto.UserResponse{
			User: dto.UserDTO{
				UserID:   userDTO.UserID,
				Username: userDTO.Username,
				TeamName: userDTO.TeamName,
				IsActive: userDTO.IsActive,
			},
		}

		common.WriteJSON(w, http.StatusOK, response)
	}
}

func GetReview(db *sql.DB) http.HandlerFunc {
	repo := NewUserRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			common.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			common.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
			return
		}

		ctx := r.Context()
		prs, err := repo.GetReviewPullRequests(ctx, userID)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get pull requests")
			return
		}

		response := map[string]interface{}{
			"user_id":       userID,
			"pull_requests": prs,
		}

		common.WriteJSON(w, http.StatusOK, response)
	}
}
