package repository

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sammidev/goca/internal/modules/note/dto"
	"github.com/sammidev/goca/internal/modules/note/entity"
	"github.com/sammidev/goca/internal/pkg/apperror"
	"github.com/sammidev/goca/internal/pkg/database"
	"github.com/sammidev/goca/internal/pkg/request"
	paging "github.com/sammidev/goca/internal/pkg/request"
)

type NotePostgresRepository struct {
	db *database.PostgreSQLDatabase
}

func NewNotePostgresRepository(db *database.PostgreSQLDatabase) *NotePostgresRepository {
	return &NotePostgresRepository{
		db: db,
	}
}

func (r *NotePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Note, error) {
	builder := sq.Select("id, user_id, url, description, created_at, updated_at").
		From("notes").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build query")
	}

	sqlExecutor, err := r.db.GetSQLExecutor(ctx)
	if err != nil {
		return nil, err
	}

	var note entity.Note
	err = sqlExecutor.QueryRow(ctx, sql, args...).Scan(
		&note.ID, &note.UserID, &note.URL, &note.Description, &note.CreatedAt, &note.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to retrieve note")
	}

	return &note, nil
}

func (r *NotePostgresRepository) Create(ctx context.Context, note *entity.Note) error {
	builder := sq.Insert("notes").Columns(
		"id", "user_id", "url", "description", "created_at", "updated_at",
	).Values(
		note.ID, note.UserID, note.URL, note.Description, note.CreatedAt, note.UpdatedAt,
	).PlaceholderFormat(sq.Dollar)

	sql, args, err := builder.ToSql()
	if err != nil {
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build query")
	}

	sqlExecutor, err := r.db.GetSQLExecutor(ctx)
	if err != nil {
		return err
	}

	_, err = sqlExecutor.Exec(ctx, sql, args...)
	if err != nil {
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to create note")
	}

	return nil
}

func (r *NotePostgresRepository) Update(ctx context.Context, note *entity.Note) error {
	builder := sq.Update("notes").
		Set("url", note.URL).
		Set("description", note.Description).
		Set("updated_at", note.UpdatedAt).
		Where(sq.Eq{"id": note.ID}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := builder.ToSql()
	if err != nil {
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build query")
	}

	sqlExecutor, err := r.db.GetSQLExecutor(ctx)
	if err != nil {
		return err
	}

	_, err = sqlExecutor.Exec(ctx, sql, args...)
	if err != nil {
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to update note")
	}

	return nil
}

func (r *NotePostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	builder := sq.Delete("notes").Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar)

	sql, args, err := builder.ToSql()
	if err != nil {
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build query")
	}

	sqlExecutor, err := r.db.GetSQLExecutor(ctx)
	if err != nil {
		return err
	}

	_, err = sqlExecutor.Exec(ctx, sql, args...)
	if err != nil {
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to delete note")
	}

	return nil
}

func (r *NotePostgresRepository) FindAll(ctx context.Context, req *dto.GetNotesRequest) (*dto.GetNotesResponse, error) {
	// 1. Buat base builder dengan SEMUA filter (WHERE clause)
	// Builder ini akan menjadi dasar untuk query count dan query data.
	baseBuilder := sq.Select(). // Select columns later
					From("notes").
					Where(sq.Eq{"user_id": req.UserID}).
					PlaceholderFormat(sq.Dollar)

	if req.HasKeyword() {
		keyword := "%" + req.Keyword + "%"
		baseBuilder = baseBuilder.Where(sq.Or{
			sq.ILike{"url": keyword},
			sq.ILike{"description": keyword},
		})
	}

	var totalData int
	// 2. Buat dan eksekusi query untuk MENGHITUNG total data
	// Gunakan baseBuilder, tapi ganti SELECT menjadi COUNT(*)
	if !req.IsUnlimitedPage() {
		countBuilder := baseBuilder.Column("COUNT(*)") // Replace select columns with COUNT(*)

		countSql, countArgs, err := countBuilder.ToSql()
		if err != nil {
			return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build count query")
		}

		sqlExecutor, err := r.db.GetSQLExecutor(ctx)
		if err != nil {
			return nil, err
		}

		// Eksekusi query count dan scan hasilnya ke variabel totalData
		err = sqlExecutor.QueryRow(ctx, countSql, countArgs...).Scan(&totalData)
		if err != nil {
			return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to execute count query")
		}
	}

	// 3. Lanjutkan builder untuk mengambil DATA aktual (tambahkan sorting & paginasi)
	dataBuilder := baseBuilder.Columns("id, user_id, url, description, created_at, updated_at") // Set columns for data fetching

	if req.HasSort() {
		dataBuilder = dataBuilder.OrderBy(req.SortBy + " " + req.SortDirection)
	}

	if !req.IsUnlimitedPage() {
		dataBuilder = dataBuilder.Limit(uint64(req.GetLimit())).Offset(uint64(req.GetOffset()))
	}

	// 4. Eksekusi query untuk mengambil data
	sql, args, err := dataBuilder.ToSql()
	if err != nil {
		return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build data query")
	}

	sqlExecutor, err := r.db.GetSQLExecutor(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := sqlExecutor.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to retrieve notes")
	}
	defer rows.Close()

	notes := make([]*entity.Note, 0)
	for rows.Next() {
		var note entity.Note
		if err := rows.Scan(&note.ID, &note.UserID, &note.URL, &note.Description, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to scan note")
		}
		notes = append(notes, &note)
	}

	// Jika tidak ada paginasi, total data adalah jumlah baris yang ditemukan
	if req.IsUnlimitedPage() {
		totalData = len(notes)
	}

	// 5. Gunakan totalData yang sudah didapat untuk membuat objek Paging
	paging, err := request.NewPaging(req.Filter.CurrentPage, req.Filter.PerPage, totalData)
	if err != nil {
		return nil, err
	}

	notesResponse := make([]*dto.NoteResponse, 0, len(notes))
	for _, note := range notes {
		notesResponse = append(notesResponse, dto.NoteEntityToNoteResponse(note))
	}

	res := dto.GetNotesResponse{
		List:   notesResponse,
		Paging: paging,
	}

	return &res, nil
}

func (r *NotePostgresRepository) Count(ctx context.Context, filter paging.Filter, userID uuid.UUID) (int, error) {
	builder := sq.Select("COUNT(*)").
		From("notes").
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar)

	if filter.HasKeyword() {
		keyword := "%" + filter.Keyword + "%"
		builder = builder.Where(sq.Or{
			sq.ILike{"url": keyword},
			sq.ILike{"description": keyword},
		})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return 0, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build query")
	}

	sqlExecutor, err := r.db.GetSQLExecutor(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = sqlExecutor.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to count notes")
	}

	return count, nil
}
