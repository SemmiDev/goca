package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/modules/note/dto"
	"github.com/sammidev/goca/internal/modules/note/entity"
	"github.com/sammidev/goca/internal/pkg/apperror"
	"github.com/sammidev/goca/internal/pkg/database"
	"github.com/sammidev/goca/internal/pkg/logger"
	"github.com/sammidev/goca/internal/pkg/validator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type NoteService struct {
	cfg       *config.Config
	logger    logger.Logger
	validator validator.Validator
	db        database.Database
	tracer    trace.Tracer
	noteRepo  NoteRepository
	UserRepo  UserRepository
}

func NewNoteService(
	cfg *config.Config,
	logger logger.Logger,
	validator validator.Validator,
	db database.Database,
	noteRepo NoteRepository,
	UserRepo UserRepository,
) *NoteService {
	return &NoteService{
		cfg:       cfg,
		logger:    logger.WithComponent("note_service"),
		validator: validator,
		db:        db,
		tracer:    otel.Tracer("note_service"),
		noteRepo:  noteRepo,
		UserRepo:  UserRepo,
	}
}

func (s *NoteService) CreateNote(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.CreateNote")
	defer span.End()

	s.logger.WithContext(ctx).Info("Creating new note", "user_id", req.UserID)
	span.SetAttributes(attribute.String("user_id", req.UserID.String()))

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewValidationError(err)
	}

	var createdNote *entity.Note
	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		note := dto.CreateNoteRequestToNoteEntity(req)

		if err := s.noteRepo.Create(txCtx, note); err != nil {
			s.logger.WithContext(txCtx).Error("Failed to create note", "error", err)
			return err
		}

		createdNote = note
		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &dto.CreateNoteResponse{
		NoteResponse: dto.NoteEntityToNoteResponse(createdNote),
	}, nil
}

func (s *NoteService) GetNote(ctx context.Context, req *dto.GetNoteRequest) (*dto.GetNoteResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetNote")
	defer span.End()

	s.logger.WithContext(ctx).Info("Retrieving note", "note_id", req.NoteID, "user_id", req.UserID)
	span.SetAttributes(
		attribute.String("note_id", req.NoteID.String()),
		attribute.String("user_id", req.UserID.String()),
	)

	note, err := s.noteRepo.GetByID(ctx, req.NoteID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrNotFound
		}
		s.logger.WithContext(ctx).Error("Failed to get note", "error", err)
		return nil, err
	}

	if err := s.checkNoteOwnership(note.UserID, req.UserID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &dto.GetNoteResponse{
		NoteResponse: dto.NoteEntityToNoteResponse(note),
	}, nil
}

func (s *NoteService) UpdateNote(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.UpdateNote")
	defer span.End()

	s.logger.WithContext(ctx).Info("Updating note", "note_id", req.NoteID, "user_id", req.UserID)
	span.SetAttributes(
		attribute.String("note_id", req.NoteID.String()),
		attribute.String("user_id", req.UserID.String()),
	)

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewValidationError(err)
	}

	var updatedNote *entity.Note
	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		note, err := s.getNoteWithOwnershipCheck(txCtx, req.NoteID, req.UserID)
		if err != nil {
			return err
		}

		req.ApplyNoteUpdates(note)

		if err := s.noteRepo.Update(txCtx, note); err != nil {
			s.logger.WithContext(txCtx).Error("Failed to update note", "error", err)
			return err
		}

		updatedNote = note
		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &dto.UpdateNoteResponse{
		NoteResponse: dto.NoteEntityToNoteResponse(updatedNote),
	}, nil
}

func (s *NoteService) DeleteNote(ctx context.Context, req *dto.DeleteNoteRequest) error {
	ctx, span := s.tracer.Start(ctx, "service.DeleteNote")
	defer span.End()

	s.logger.WithContext(ctx).Info("Deleting note", "note_id", req.NoteID, "user_id", req.UserID)
	span.SetAttributes(
		attribute.String("note_id", req.NoteID.String()),
		attribute.String("user_id", req.UserID.String()),
	)

	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		if _, err := s.getNoteWithOwnershipCheck(txCtx, req.NoteID, req.UserID); err != nil {
			return err
		}

		if err := s.noteRepo.Delete(txCtx, req.NoteID); err != nil {
			s.logger.WithContext(txCtx).Error("Failed to delete note", "error", err)
			return err
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (s *NoteService) GetNotes(ctx context.Context, req *dto.GetNotesRequest) (*dto.GetNotesResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetNotes")
	defer span.End()

	s.logger.WithContext(ctx).Info("Listing notes", "user_id", req.UserID)
	span.SetAttributes(attribute.String("user_id", req.UserID.String()))

	res, err := s.noteRepo.FindAll(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.logger.WithContext(ctx).Error("Failed to find notes", "error", err)
		return nil, err
	}

	return res, nil
}

// Helper methods

func (s *NoteService) checkNoteOwnership(noteUserID, requestUserID uuid.UUID) error {
	if noteUserID != requestUserID {
		return apperror.ErrForbidden
	}
	return nil
}

func (s *NoteService) getNoteWithOwnershipCheck(ctx context.Context, noteID, userID uuid.UUID) (*entity.Note, error) {
	ctx, span := s.tracer.Start(ctx, "helper.getNoteWithOwnershipCheck")
	defer span.End()

	span.SetAttributes(
		attribute.String("note_id", noteID.String()),
		attribute.String("user_id", userID.String()),
	)

	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrNotFound
		}
		s.logger.WithContext(ctx).Error("Failed to get note", "error", err)
		return nil, apperror.ErrInternalError
	}

	if err := s.checkNoteOwnership(note.UserID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return note, nil
}
