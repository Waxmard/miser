package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Waxmard/miser/internal/repository"
)

type rawEmailRepo struct {
	db *sql.DB
}

func (r *rawEmailRepo) Create(ctx context.Context, email *repository.RawEmail) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO raw_emails (id, message_id, subject, sender, body, received_at, status, error, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		email.ID, email.MessageID, email.Subject, email.From, email.Body,
		email.ReceivedAt.Format(timeFormat), email.Status, email.Error,
		email.CreatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("insert raw email: %w", err)
	}
	return nil
}

func (r *rawEmailRepo) GetPending(ctx context.Context, limit int) ([]repository.RawEmail, error) {
	query := `SELECT id, message_id, subject, sender, body, received_at, status, error, created_at
		 FROM raw_emails WHERE status = 'pending' ORDER BY received_at ASC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get pending emails: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var emails []repository.RawEmail
	for rows.Next() {
		var e repository.RawEmail
		var receivedAt, createdAt string
		if err := rows.Scan(&e.ID, &e.MessageID, &e.Subject, &e.From, &e.Body,
			&receivedAt, &e.Status, &e.Error, &createdAt); err != nil {
			return nil, fmt.Errorf("scan raw email: %w", err)
		}
		e.ReceivedAt = parseTime(receivedAt)
		e.CreatedAt = parseTime(createdAt)
		emails = append(emails, e)
	}
	return emails, rows.Err()
}

func (r *rawEmailRepo) MarkProcessed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE raw_emails SET status = 'processed' WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("mark email processed: %w", err)
	}
	return nil
}

func (r *rawEmailRepo) MarkFailed(ctx context.Context, id, reason string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE raw_emails SET status = 'failed', error = ? WHERE id = ?`, reason, id)
	if err != nil {
		return fmt.Errorf("mark email failed: %w", err)
	}
	return nil
}
