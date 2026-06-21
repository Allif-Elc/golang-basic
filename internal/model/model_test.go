package model

import (
	"testing"
)

func TestPageResult(t *testing.T) {
	pr := PageResult[string]{
		Data:       []string{"a", "b"},
		Page:       1,
		Size:       2,
		StartRow:   1,
		EndRow:     2,
		NextCursor: 0,
	}
	if len(pr.Data) != 2 {
		t.Errorf("Data length = %d, want 2", len(pr.Data))
	}
	if pr.Page != 1 {
		t.Errorf("Page = %d, want 1", pr.Page)
	}
}

func TestPageResult_Empty(t *testing.T) {
	pr := PageResult[int]{}
	if pr.Data != nil {
		t.Error("Data should be nil for zero value")
	}
	if pr.Page != 0 {
		t.Errorf("Page = %d, want 0", pr.Page)
	}
}

func TestPageRequest(t *testing.T) {
	pr := PageRequest{Page: 1, Size: 10}
	if pr.Page != 1 || pr.Size != 10 {
		t.Error("PageRequest fields mismatch")
	}
}

func TestPageRequest_ZeroValues(t *testing.T) {
	pr := PageRequest{}
	if pr.Page != 0 || pr.Size != 0 {
		t.Error("PageRequest should have zero values")
	}
}
