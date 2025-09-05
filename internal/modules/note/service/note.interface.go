package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/modules/note/dto"
	"github.com/sammidev/goca/internal/modules/note/entity"
	userEntity "github.com/sammidev/goca/internal/modules/user/entity"
)

type NoteRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Note, error)
	Create(ctx context.Context, note *entity.Note) error
	Update(ctx context.Context, note *entity.Note) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context, req *dto.GetNotesRequest) (*dto.GetNotesResponse, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error)
}
