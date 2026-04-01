package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Waxmard/miser/internal/repository"
)

type syncStateRepo struct {
	db *sql.DB
}

func (r *syncStateRepo) Get(ctx context.Context, source string) (*repository.SyncState, error) {
	var s repository.SyncState
	var lastSyncAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT source, last_sync_at, last_message_uid, metadata FROM sync_state WHERE source = ?`, source,
	).Scan(&s.Source, &lastSyncAt, &s.LastMessageUID, &s.Metadata)
	if err != nil {
		return nil, fmt.Errorf("get sync state: %w", err)
	}
	s.LastSyncAt = parseTime(lastSyncAt)
	return &s, nil
}

func (r *syncStateRepo) Upsert(ctx context.Context, s *repository.SyncState) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sync_state (source, last_sync_at, last_message_uid, metadata)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(source) DO UPDATE SET last_sync_at = ?, last_message_uid = ?, metadata = ?`,
		s.Source, s.LastSyncAt.Format(timeFormat), s.LastMessageUID, s.Metadata,
		s.LastSyncAt.Format(timeFormat), s.LastMessageUID, s.Metadata,
	)
	if err != nil {
		return fmt.Errorf("upsert sync state: %w", err)
	}
	return nil
}
