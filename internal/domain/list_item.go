package domain

import (
	"errors"
	"time"
)

type ListItemType string

const (
	ListItemTypeTask ListItemType = "task"
	ListItemTypeDone ListItemType = "done"
)

var validListItemTypes = map[ListItemType]string{
	ListItemTypeTask: ".",
	ListItemTypeDone: "x",
}

func (t ListItemType) IsValid() bool {
	_, ok := validListItemTypes[t]
	return ok
}

func (t ListItemType) Symbol() string {
	return validListItemTypes[t]
}

type ListItem struct {
	VersionInfo
	ListEntityID EntityID
	Type         ListItemType
	Content      string
	CreatedAt    time.Time
}

func NewListItem(listEntityID EntityID, itemType ListItemType, content string) ListItem {
	return ListItem{
		VersionInfo: VersionInfo{
			EntityID: NewEntityID(),
			Version:  1,
			ValidFrom: time.Now(),
			OpType:   OpTypeInsert,
		},
		ListEntityID: listEntityID,
		Type:         itemType,
		Content:      content,
		CreatedAt:    time.Now(),
	}
}

func (li ListItem) Validate() error {
	if li.ListEntityID.IsEmpty() {
		return errors.New("list entity ID is required")
	}
	if !li.Type.IsValid() {
		return errors.New("invalid list item type")
	}
	if li.Content == "" {
		return errors.New("content cannot be empty")
	}
	return nil
}

func (li ListItem) IsComplete() bool {
	return li.Type == ListItemTypeDone
}
