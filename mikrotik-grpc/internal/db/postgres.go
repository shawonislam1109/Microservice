package db

import (
	"context"
	"errors"
	"fmt"
	"isp-management-system/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for all database operations.
type Repository interface {
	// User methods
	CreateUser(ctx context.Context, user *models.User, packageName string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)

	// Session methods
	CreateSession(ctx context.Context, session *models.Session) error
	UpdateSession(ctx context.Context, session *models.Session) error

	// General
	Close()
}

// postgresRepository implements the Repository interface for PostgreSQL.
type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository.
func NewPostgresRepository(connString string) (Repository, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &postgresRepository{pool: pool}, nil
}

// CreateUser finds a package by name and creates a new user.
func (r *postgresRepository) CreateUser(ctx context.Context, user *models.User, packageName string) (*models.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback is a no-op if tx has been committed

	// 1. Get Package ID from package name
	var packageID string
	err = tx.QueryRow(ctx, "SELECT id FROM packages WHERE name = $1", packageName).Scan(&packageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("package '%s' not found", packageName)
		}
		return nil, fmt.Errorf("failed to query package: %w", err)
	}
	user.PackageID = packageID

	// 2. Insert the new user
	query := `
		INSERT INTO users (username, password, package_id, status, expiry_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, query, user.Username, user.Password, user.PackageID, user.Status, user.ExpiryDate).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// Check for unique violation for username
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fmt.Errorf("username '%s' already exists", user.Username)
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

// GetUserByUsername fetches a user and their package information by username.
func (r *postgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.password, u.package_id, u.status, u.expiry_date, p.rate_limit
		FROM users u
		JOIN packages p ON u.package_id = p.id
		WHERE u.username = $1`

	var user models.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.PackageID, &user.Status, &user.ExpiryDate, &user.PackageRateLimit,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query row scan failed: %w", err)
	}
	return &user, nil
}

// CreateSession inserts a new accounting session (typically for Acct-Start).
func (r *postgresRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (session_id, username, nas_ip_address, calling_station_id, session_start_time, input_octets, output_octets)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		session.SessionID, session.Username, session.NASIPAddress, session.CallingStationID, session.SessionStartTime, session.InputOctets, session.OutputOctets,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// UpdateSession updates an existing session (for Acct-Interim-Update or Acct-Stop).
func (r *postgresRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE sessions
		SET
			session_stop_time = $1,
			session_total_time = $2,
			input_octets = $3,
			output_octets = $4,
			terminate_cause = $5
		WHERE session_id = $6 AND username = $7`

	var stopTime *time.Time
	if !session.SessionStopTime.IsZero() {
		stopTime = &session.SessionStopTime
	}

	_, err := r.pool.Exec(ctx, query,
		stopTime, session.SessionTotalTime, session.InputOctets, session.OutputOctets, session.TerminateCause, session.SessionID, session.Username,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// Close closes the database connection pool.
func (r *postgresRepository) Close() {
	r.pool.Close()
}