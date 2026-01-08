package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOpType_Insert_IsValid(t *testing.T) {
	assert.True(t, OpTypeInsert.IsValid())
}

func TestOpType_Update_IsValid(t *testing.T) {
	assert.True(t, OpTypeUpdate.IsValid())
}

func TestOpType_Delete_IsValid(t *testing.T) {
	assert.True(t, OpTypeDelete.IsValid())
}

func TestOpType_Invalid_IsNotValid(t *testing.T) {
	invalid := OpType("invalid")
	assert.False(t, invalid.IsValid())
}

func TestOpType_String_ReturnsValue(t *testing.T) {
	assert.Equal(t, "INSERT", OpTypeInsert.String())
	assert.Equal(t, "UPDATE", OpTypeUpdate.String())
	assert.Equal(t, "DELETE", OpTypeDelete.String())
}

func TestVersionInfo_IsCurrent_WhenValidToNil_ReturnsTrue(t *testing.T) {
	v := VersionInfo{ValidTo: nil}

	assert.True(t, v.IsCurrent())
}

func TestVersionInfo_IsCurrent_WhenValidToSet_ReturnsFalse(t *testing.T) {
	now := time.Now()
	v := VersionInfo{ValidTo: &now}

	assert.False(t, v.IsCurrent())
}

func TestVersionInfo_IsDeleted_WhenOpTypeDelete_ReturnsTrue(t *testing.T) {
	v := VersionInfo{OpType: OpTypeDelete}

	assert.True(t, v.IsDeleted())
}

func TestVersionInfo_IsDeleted_WhenOpTypeInsert_ReturnsFalse(t *testing.T) {
	v := VersionInfo{OpType: OpTypeInsert}

	assert.False(t, v.IsDeleted())
}

func TestVersionInfo_IsDeleted_WhenOpTypeUpdate_ReturnsFalse(t *testing.T) {
	v := VersionInfo{OpType: OpTypeUpdate}

	assert.False(t, v.IsDeleted())
}
