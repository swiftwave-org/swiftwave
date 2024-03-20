package core

import (
	"context"
	"gorm.io/gorm"
)

func (s *ServerResourceStat) Create(_ context.Context, db gorm.DB) error {
	return db.Create(s).Error
}
