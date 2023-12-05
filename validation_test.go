package db

import (
	"errors"
	"testing"
)

var nameErr = errors.New("the name field is required")
var nameMatchErr = errors.New("the name field doesn't match")

func (p *Parent) Validate() error {
	err := Require(p, "parent_name")
	if err != nil {
		return nameErr
	}

	err = Match(p, "^[A-Za-z ]+$", "parent_name")
	if err != nil {
		return nameMatchErr
	}

	return nil
}

func (p *Parent) ValidateInsert() error {
	if p.ID == 0 {
		p.ID = 456
	}
	return nil
}

func (p *Parent) Filter(fields map[string]any) error {
	delete(fields, "no_field")
	return nil
}

func TestValidate(t *testing.T) {
	err := doValidateInsert(&Parent{})
	if !errors.Is(err, nameErr) {
		t.Error("Should have an invalid name")
	}

	err = doValidateInsert(&Parent{Name: NewNullable("Name")})
	if err != nil {
		t.Error("Should validate")
	}
}

func TestValidateInsert(t *testing.T) {
	err := doValidateUpdate(&Parent{Name: NewNullable("1 Invalid")})
	if err == nil {
		t.Error("Should not validate")
	}

	err = doValidateInsert(&Parent{Name: NewNullable("Name")})
	if err != nil {
		t.Error("Should validate", err)
	}
}

func TestValidateUpdate(t *testing.T) {
	err := doValidateUpdate(&Parent{ID: 555, Name: NewNullable("Name")})
	if err != nil {
		t.Error("Should validate", err)
	}

	err = doValidateUpdate(&Parent{Name: NewNullable("123")})
	if err == nil {
		t.Error("Should not validate")
	}
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
