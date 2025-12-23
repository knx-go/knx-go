package pgknx

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	db *gorm.DB
}

func New(dsn string, debug bool) (*Store, error) {
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	if debug {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	var err error
	s := Store{}
	if s.db, err = gorm.Open(postgres.Open(dsn), gormConfig); err != nil {
		return nil, err
	}
	if err = s.db.AutoMigrate(&Event{}); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Store) Close() {
}

func (s *Store) CreateEvent(ctx context.Context, e Event) error {
	return gorm.G[Event](s.db).Create(ctx, &e)
}

func (s *Store) GroupLastEvent(ctx context.Context, group string) (Event, error) {
	return gorm.G[Event](s.db).Where("group = ?", group).Order("timestamp desc").Take(ctx)
}
