package core

import (
	"context"
	"errors"
	"github.com/oklog/ulid"
	"gorm.io/gorm"
	"math/rand"
	"strings"
	"time"
)

func (l *AppBasicAuthAccessControlList) Create(_ context.Context, db gorm.DB) error {
	l.Name = strings.TrimSpace(l.Name)
	if strings.Compare(l.Name, "") == 0 {
		return errors.New("name cannot be empty")
	}
	l.GeneratedName = ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
	return db.Create(&l).Error
}

func (l *AppBasicAuthAccessControlList) Delete(_ context.Context, db gorm.DB) error {
	return db.Delete(&l).Error
}
