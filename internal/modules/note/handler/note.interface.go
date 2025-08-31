package handler

import (
	"context"

	"github.com/sammidev/goca/internal/modules/note/dto"
)

type NoteService interface {
	GetNotes(ctx context.Context, req *dto.GetNotesRequest) (*dto.GetNotesResponse, error)
	CreateNote(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error)
	GetNote(ctx context.Context, req *dto.GetNoteRequest) (*dto.GetNoteResponse, error)
	UpdateNote(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error)
	DeleteNote(ctx context.Context, req *dto.DeleteNoteRequest) error
}
