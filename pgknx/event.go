package pgknx

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Timestamp   time.Time `gorm:"index"`
	Command     string
	Source      string
	Destination string
	Data        []byte
	Group       string `gorm:"index"`
	DPT         string
	Decoded     string
}

func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New()
	return nil
}

func (s *Store) CreateEvent(ctx context.Context, e Event) error {
	return gorm.G[Event](s.db).Create(ctx, &e)
}

func (s *Store) GroupLastEvent(ctx context.Context, group string) (Event, error) {
	return gorm.G[Event](s.db).Where("\"group\" = ?", group).Order("timestamp desc").Take(ctx)
}

type HistoryOptions struct {
	From       time.Time
	To         time.Time
	Limit      int
	Offset     int
	Descending bool
}

func (s *Store) GroupEventsHistory(ctx context.Context, group string, opts HistoryOptions) ([]Event, error) {
	var events []Event
	query := s.db.Where("\"group\" = ?", group)

	if !opts.From.IsZero() {
		query.Where("timestamp >= ?", opts.From)
	}
	if !opts.To.IsZero() {
		query.Where("timestamp <= ?", opts.To)
	}
	if opts.Descending {
		query.Order("timestamp desc")
	}
	if opts.Limit > 0 {
		query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query.Offset(opts.Offset)
	}

	results := query.Find(&events)
	return events, results.Error
}
