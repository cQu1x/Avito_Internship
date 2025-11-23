package handlers

import (
	"AvitoInternship/internal/handlers/pullRequest"
	"AvitoInternship/internal/handlers/team"
	"AvitoInternship/internal/handlers/user"
	"database/sql"
	"net/http"
)

func SetupRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/users/setIsActive", user.SetIsActive(db))
	mux.HandleFunc("/users/getReview", user.GetReview(db))

	mux.HandleFunc("/team/add", team.AddTeam(db))
	mux.HandleFunc("/team/get", team.GetTeam(db))

	mux.HandleFunc("/pullRequest/create", pullRequest.Create(db))
	mux.HandleFunc("/pullRequest/merge", pullRequest.Merge(db))
	mux.HandleFunc("/pullRequest/reassign", pullRequest.Reassign(db))

	return mux
}
