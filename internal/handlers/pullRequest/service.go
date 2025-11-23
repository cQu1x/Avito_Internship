package pullRequest

import (
	"context"
	"database/sql"
	"errors"

	dto "AvitoInternship/internal/handlers/dto"
)

type PullRequestService struct {
	prRepo   *PullRequestRepository
	userRepo UserRepo
}

func NewPullRequestService(prRepo *PullRequestRepository, userRepo UserRepo) *PullRequestService {
	return &PullRequestService{prRepo: prRepo, userRepo: userRepo}
}

type UserRepo interface {
	GetByID(ctx context.Context, userID string) (*dto.UserDTO, error)
	GetRandomActiveTeammate(ctx context.Context, teamID int, exclude []string) (*dto.UserDTO, error)
}

func (s *PullRequestService) reassignReviewerTx(ctx context.Context, prID string, oldReviewerID string, out **dto.ReassignReviewerResponse) error {
	tx, err := s.prRepo.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	pr, err := s.prRepo.GetByIDForUpdateTx(ctx, tx, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(dto.ErrorCodeNotFound)
		}
		return err
	}

	if pr.Status == "MERGED" {
		return errors.New(dto.ErrorCodePRMerged)
	}

	if !pr.HasReviewer(oldReviewerID) {
		return errors.New(dto.ErrorCodeNotFound)
	}

	oldUser, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(dto.ErrorCodeNotFound)
		}
		return err
	}
	exclude := append(pr.ReviewerIDs[:0:0], pr.ReviewerIDs...)
	exclude = append(exclude, oldReviewerID)
	exclude = append(exclude, pr.AuthorID)
	replacement, err := s.userRepo.GetRandomActiveTeammate(ctx, oldUser.TeamID, exclude)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(dto.ErrorCodeNoCandidate)
		}
		return err
	}
	if replacement.UserID == pr.AuthorID {
		return errors.New(dto.ErrorCodeNoCandidate)
	}

	if err := s.prRepo.ReassignReviewerTx(ctx, tx, pr.ID, oldReviewerID, replacement.UserID); err != nil {
		return err
	}

	reviewers, err := s.prRepo.ListReviewersTx(ctx, tx, pr.ID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	*out = &dto.ReassignReviewerResponse{
		PR: dto.PullRequestDTO{
			PullRequestID:     pr.ID,
			PullRequestName:   pr.Title,
			AuthorID:          pr.AuthorID,
			Status:            pr.Status,
			AssignedReviewers: reviewers,
		},
		ReplacedBy: replacement.UserID,
	}
	return nil
}

func (s *PullRequestService) Reassign(ctx context.Context, prID, oldReviewerID string) (*dto.ReassignReviewerResponse, error) {
	var out *dto.ReassignReviewerResponse
	if err := s.reassignReviewerTx(ctx, prID, oldReviewerID, &out); err != nil {
		return nil, err
	}
	return out, nil
}
