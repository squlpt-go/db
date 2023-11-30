package db

import (
	"errors"
	"testing"
)

var nameErr = errors.New("the name field is required")
var nameMatchErr = errors.New("the name field doesn't match")

func init() {
	db := DB()
	Validate(db,
		Rule[Parent](
			func(entity *Parent) error {
				if entity.ID == 0 {
					entity.ID = 456
				}
				return nil
			},
		).ForInsert(),
		Required[Parent]("parent_name").WithErr(nameErr),
		Matches[Parent]("^[A-Za-z ]+$", "parent_name").WithErr(nameMatchErr),
		Filter[Parent](
			func(fields map[string]any) {
				delete(fields, "no_field")
			},
		),
	)
}

func TestValidate(t *testing.T) {
	db := DB()

	err := DoValidateInsert(db, &Parent{})
	if !errors.Is(err, nameErr) {
		t.Error("Should have an invalid name")
	}

	err = DoValidateInsert(db, &Parent{Name: NewNullable("Name")})
	if err != nil {
		t.Error("Should validate")
	}
}

func TestValidateInsert(t *testing.T) {
	db := DB()

	err := DoValidateUpdate(db, &Parent{Name: NewNullable("1 Invalid")})
	if err == nil {
		t.Error("Should not validate")
	}

	err = DoValidateInsert(db, &Parent{Name: NewNullable("Name")})
	if err != nil {
		t.Error("Should validate", err)
	}
}

func TestValidateUpdate(t *testing.T) {
	db := DB()

	err := DoValidateUpdate(db, &Parent{ID: 555, Name: NewNullable("Name")})
	if err != nil {
		t.Error("Should validate", err)
	}

	err = DoValidateUpdate(db, &Parent{Name: NewNullable("123")})
	if err == nil {
		t.Error("Should not validate")
	}
}

func TestFilterInsert(t *testing.T) {
	db := DB()

	v := map[string]any{
		"no_field": 1,
	}

	DoFilterInsert[Parent](db, v)
	_, has := v["no_field"]

	if has {
		t.Error("Should not have 'no_field'")
	}
}

func TestFilterUpdate(t *testing.T) {
	db := DB()

	v := map[string]any{
		"no_field": 1,
	}

	DoFilterInsert[Parent](db, v)
	_, has := v["no_field"]

	if has {
		t.Error("Should not have 'no_field'")
	}
}
