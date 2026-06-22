package postgres

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/history"
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // register "pgx" driver for database/sql
)

const QSaveCommandRuns = `
INSERT INTO command_runs (command,
profile, status, output, error_text)
VALUES ($1, $2, $3, $4, $5)
`

const QListCommandRuns = `
SELECT id, command, profile, status, output, error_text, created_at
FROM command_runs
ORDER BY created_at DESC
LIMIT $1
`

const QInspectSnapshot = `
INSERT INTO inspect_snapshots (app_name, profile, payload_json)
VALUES ($1, $2, $3)
`

const QLatestsSnapshot = `
SELECT id, app_name, profile, payload_json, created_at
FROM inspect_snapshots
WHERE app_name = $1
ORDER BY created_at DESC
LIMIT 1
`

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) SaveCommandRun(ctx context.Context, run history.CommandRun) error {
	_, err := s.db.ExecContext(ctx, QSaveCommandRuns,
		run.Command, run.Profile, run.Status, run.Output, run.ErrorText,
	)
	return err
}

func (s *Store) ListCommandRuns(ctx context.Context, limit int) ([]history.CommandRun, error) {
	rows, err := s.db.QueryContext(ctx, QListCommandRuns, limit)
	if err != nil {
		return nil, fmt.Errorf("query list command runs: %w", err)
	}
	defer rows.Close()

	var runs []history.CommandRun
	for rows.Next() {
		var r history.CommandRun
		if err := rows.Scan(
			&r.ID,
			&r.Command,
			&r.Profile,
			&r.Status,
			&r.Output,
			&r.ErrorText,
			&r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan command run: %w", err)
		}
		runs = append(runs, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows scan: %w", err)
	}
	return runs, nil
}

func (s *Store) SaveInspectSnapshot(ctx context.Context, snap history.InspectSnapshot) error {
	_, err := s.db.ExecContext(ctx, QInspectSnapshot,
		snap.AppName, snap.Profile, snap.PayloadJSON,
	)
	return err
}

func (s *Store) LatestSnapshot(ctx context.Context, appName string) (history.InspectSnapshot, error) {
	var snap history.InspectSnapshot
	err := s.db.QueryRowContext(ctx, QLatestsSnapshot, appName).Scan(
		&snap.ID,
		&snap.AppName,
		&snap.Profile,
		&snap.PayloadJSON,
		&snap.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return snap, fmt.Errorf("%w: no snapshot for app %q", domain.ErrNotFound, appName)
		}
		return snap, fmt.Errorf("query latest snapshot for app: %w", err)
	}
	return snap, nil
}
