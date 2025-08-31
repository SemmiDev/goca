package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/modules/note/entity"
	"github.com/sammidev/goca/internal/pkg/request"
)

type NoteResponse struct {
	ID          uuid.UUID `json:"id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	UserID      uuid.UUID `json:"user_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	URL         string    `json:"url" example:"https://example.com"`
	Description string    `json:"description" example:"This is a note description"`
	CreatedAt   time.Time `json:"created_at" example:"2025-06-01T20:50:35.388851+07:00"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-06-01T20:50:35.388851+07:00"`
}

type CreateNoteRequest struct {
	UserID      uuid.UUID `json:"user_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	URL         string    `json:"url" validate:"required,url" example:"https://example.com"`
	Description string    `json:"description" validate:"required,min=5" example:"This is a note description"`
}

type GetNoteRequest struct {
	UserID uuid.UUID `json:"user_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	NoteID uuid.UUID `json:"note_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
}

type GetNoteResponse struct {
	*NoteResponse
}

type CreateNoteResponse struct {
	*NoteResponse
}

type UpdateNoteRequest struct {
	UserID      uuid.UUID `json:"user_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	NoteID      uuid.UUID `json:"note_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	URL         *string   `json:"url" validate:"omitempty,url" example:"https://updated-example.com"`
	Description *string   `json:"description" validate:"omitempty,min=5" example:"Updated note description"`
}

func (u UpdateNoteRequest) ApplyNoteUpdates(note *entity.Note) {
	if u.URL != nil {
		note.URL = *u.URL
	}

	if u.Description != nil {
		note.Description = *u.Description
	}

	note.UpdatedAt = time.Now()
}

type UpdateNoteResponse struct {
	*NoteResponse
}

type DeleteNoteRequest struct {
	UserID uuid.UUID `json:"user_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	NoteID uuid.UUID `json:"note_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
}

type GetNotesRequest struct {
	UserID uuid.UUID `json:"user_id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	request.Filter
}

func NewGetNotesRequest() *GetNotesRequest {
	return &GetNotesRequest{
		Filter: request.NewFilter(),
	}
}

type GetNotesResponse struct {
	List   []*NoteResponse `json:"list"`
	Paging *request.Paging `json:"paging"`
}

func CreateNoteRequestToNoteEntity(payload *CreateNoteRequest) *entity.Note {
	return &entity.Note{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      payload.UserID,
		URL:         payload.URL,
		Description: payload.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NoteEntityToNoteResponse(note *entity.Note) *NoteResponse {
	return &NoteResponse{
		ID:          note.ID,
		UserID:      note.UserID,
		URL:         note.URL,
		Description: note.Description,
		CreatedAt:   note.CreatedAt,
		UpdatedAt:   note.UpdatedAt,
	}
}

func NotesEntityToNotesResponse(notes []*entity.Note) []*NoteResponse {
	responses := make([]*NoteResponse, len(notes))
	for i, note := range notes {
		responses[i] = NoteEntityToNoteResponse(note)
	}
	return responses
}
