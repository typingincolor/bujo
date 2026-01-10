package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOpType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		opType   OpType
		expected bool
	}{
		{"INSERT is valid", OpTypeInsert, true},
		{"UPDATE is valid", OpTypeUpdate, true},
		{"DELETE is valid", OpTypeDelete, true},
		{"invalid op type", OpType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.opType.IsValid())
		})
	}
}

func TestOpType_String(t *testing.T) {
	assert.Equal(t, "INSERT", OpTypeInsert.String())
	assert.Equal(t, "UPDATE", OpTypeUpdate.String())
	assert.Equal(t, "DELETE", OpTypeDelete.String())
}

func TestVersionInfo_IsCurrent(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		info     VersionInfo
		expected bool
	}{
		{"nil ValidTo is current", VersionInfo{ValidTo: nil}, true},
		{"set ValidTo is not current", VersionInfo{ValidTo: &now}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.info.IsCurrent())
		})
	}
}

func TestVersionInfo_IsDeleted(t *testing.T) {
	tests := []struct {
		name     string
		opType   OpType
		expected bool
	}{
		{"DELETE op type is deleted", OpTypeDelete, true},
		{"INSERT op type is not deleted", OpTypeInsert, false},
		{"UPDATE op type is not deleted", OpTypeUpdate, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := VersionInfo{OpType: tt.opType}
			assert.Equal(t, tt.expected, v.IsDeleted())
		})
	}
}
