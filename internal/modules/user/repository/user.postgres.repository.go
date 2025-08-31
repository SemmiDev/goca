package repo

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sammidev/goca/internal/modules/user/entity"
	"github.com/sammidev/goca/internal/pkg/apperror"
	"github.com/sammidev/goca/internal/pkg/database"
)

type UserPostgresRepository struct {
	db *database.PostgreSQLDatabase
}

func NewUserPostgresRepository(db *database.PostgreSQLDatabase) *UserPostgresRepository {
	return &UserPostgresRepository{
		db: db,
	}
}

func (r *UserPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	builder := sq.Select("id, email, email_verified_at, first_name, last_name, full_name, password, status, two_factor_secret, two_factor_enabled, created_at, updated_at").
		From("users").
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

	var user entity.User
	err = sqlExecutor.QueryRow(ctx, sql, args...).Scan(
		&user.ID, &user.Email, &user.EmailVerifiedAt, &user.FirstName, &user.LastName, &user.FullName,
		&user.Password, &user.Status, &user.TwoFactorSecret, &user.TwoFactorEnabled,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to retrieve user")
	}

	return &user, nil
}

func (r *UserPostgresRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	builder := sq.Select("id, email, email_verified_at, first_name, last_name, full_name, password, status, two_factor_secret, two_factor_enabled, created_at, updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to build query")
	}

	sqlExecutor, err := r.db.GetSQLExecutor(ctx)
	if err != nil {
		return nil, err
	}

	var user entity.User
	err = sqlExecutor.QueryRow(ctx, sql, args...).Scan(
		&user.ID, &user.Email, &user.EmailVerifiedAt, &user.FirstName, &user.LastName, &user.FullName,
		&user.Password, &user.Status, &user.TwoFactorSecret, &user.TwoFactorEnabled,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to retrieve user")
	}

	return &user, nil
}

func (r *UserPostgresRepository) Create(ctx context.Context, user *entity.User) error {
	builder := sq.Insert("users").Columns(
		"id", "email", "email_verified_at", "first_name", "last_name", "full_name", "password", "status", "two_factor_secret", "two_factor_enabled", "created_at", "updated_at",
	).Values(
		user.ID, user.Email, user.EmailVerifiedAt, user.FirstName,
		user.LastName, user.FullName, user.Password, user.Status, user.TwoFactorSecret, user.TwoFactorEnabled,
		user.CreatedAt, user.UpdatedAt,
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
		if database.IsUniqueViolation(err) {
			return apperror.ErrUserAlreadyExists
		}
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to create user")
	}

	return nil
}

func (r *UserPostgresRepository) Update(ctx context.Context, user *entity.User) error {
	builder := sq.Update("users").
		Set("first_name", user.FirstName).
		Set("email", user.Email).
		Set("email_verified_at", user.EmailVerifiedAt).
		Set("last_name", user.LastName).
		Set("full_name", user.FullName).
		Set("password", user.Password).
		Set("status", user.Status).
		Set("two_factor_secret", user.TwoFactorSecret).
		Set("two_factor_enabled", user.TwoFactorEnabled).
		Set("updated_at", user.UpdatedAt).
		Where(sq.Eq{"id": user.ID}).PlaceholderFormat(sq.Dollar)

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
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to update user")
	}

	return nil
}

func (r *UserPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	builder := sq.Delete("users").Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar)

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
		return apperror.WrapError(err, apperror.ErrCodeDatabaseError, "Failed to delete user")
	}

	return nil
}
