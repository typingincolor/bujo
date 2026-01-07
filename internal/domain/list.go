package domain

import (
	"errors"
	"strings"
	"time"
)

type List struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

func NewList(name string) List {
	return List{
		Name:      strings.TrimSpace(name),
		CreatedAt: time.Now(),
	}
}

func (l List) Validate() error {
	if strings.TrimSpace(l.Name) == "" {
		return errors.New("list name cannot be empty")
	}
	return nil
}
