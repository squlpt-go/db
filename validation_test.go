package db

import (
	"testing"
)

func (p *Parent) Filter(fields map[string]any) error {
	delete(fields, "no_field")
	return nil
}

func TestFilterInsert(t *testing.T) {
	v, err := doFilterInsert[Parent](&Parent{DontSave: "this will be removed"})
	if err != nil {
		t.Error("Should validate", err)
	}
	_, has := v["no_field"]

	if has {
		t.Error("Should not have 'no_field'")
	}
}

func TestFilterUpdate(t *testing.T) {
	v, err := doFilterUpdate[Parent](&Parent{DontSave: "this will be removed"})
	if err != nil {
		t.Error("Should validate", err)
	}
	_, has := v["no_field"]

	if has {
		t.Error("Should not have 'no_field'")
	}
}
