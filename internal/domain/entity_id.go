package domain

import (
	"errors"

	"github.com/google/uuid"
)

type EntityID string

func NewEntityID() EntityID {
	return EntityID(uuid.New().String())
}

func ParseEntityID(s string) (EntityID, error) {
	if s == "" {
		return "", errors.New("entity ID cannot be empty")
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return "", errors.New("invalid entity ID format")
	}
	return EntityID(parsed.String()), nil
}

func (e EntityID) String() string {
	return string(e)
}

func (e EntityID) IsEmpty() bool {
	return e == ""
}
