package domain

import "time"

type OpType string

const (
	OpTypeInsert OpType = "INSERT"
	OpTypeUpdate OpType = "UPDATE"
	OpTypeDelete OpType = "DELETE"
)

var validOpTypes = map[OpType]bool{
	OpTypeInsert: true,
	OpTypeUpdate: true,
	OpTypeDelete: true,
}

func (o OpType) IsValid() bool {
	return validOpTypes[o]
}

func (o OpType) String() string {
	return string(o)
}

type VersionInfo struct {
	RowID     int64
	EntityID  EntityID
	Version   int
	ValidFrom time.Time
	ValidTo   *time.Time
	OpType    OpType
}

func (v VersionInfo) IsCurrent() bool {
	return v.ValidTo == nil
}

func (v VersionInfo) IsDeleted() bool {
	return v.OpType == OpTypeDelete
}
