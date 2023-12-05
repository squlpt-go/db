package db

import "testing"

func TestRaw_ValidateSQL(t *testing.T) {
	r := Raw("SELECT * FROM table")
	err := r.ValidateSQL()
	if err != nil {
		t.Errorf("Raw.ValidateSQL() returned an error: %v", err)
	}

	r = Raw("")
	err = r.ValidateSQL()
	if err == nil {
		t.Error("Raw.ValidateSQL() should return an error for empty raw SQL")
	}
}

func TestIdent_ValidateSQL(t *testing.T) {
	i := Ident("name")
	err := i.ValidateSQL()
	if err != nil {
		t.Errorf("Ident.ValidateSQL() returned an error: %v", err)
	}

	i = Ident("")
	err = i.ValidateSQL()
	if err == nil {
		t.Error("Ident.ValidateSQL() should return an error for empty ident")
	}
}

func TestString_ValidateSQL(t *testing.T) {
	s := String("value")
	err := s.ValidateSQL()
	if err != nil {
		t.Errorf("String.ValidateSQL() returned an error: %v", err)
	}
}

func TestInt_ValidateSQL(t *testing.T) {
	i := Int(123)
	err := i.ValidateSQL()
	if err != nil {
		t.Errorf("Int.ValidateSQL() returned an error: %v", err)
	}
}

func TestQuery_ValidateSQL(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	err := q.ValidateSQL()
	if err != nil {
		t.Errorf("ValidateSQL() returned an error for a valid query: %v", err)
	}
}

func TestRaw_ValidateSQL_Empty(t *testing.T) {
	r := Raw("")

	err := r.ValidateSQL()
	if err == nil {
		t.Error("Raw.ValidateSQL() should return an error for empty raw SQL")
	}
}

func TestIdent_ValidateSQL_Empty(t *testing.T) {
	i := Ident("")

	err := i.ValidateSQL()
	if err == nil {
		t.Error("Ident.ValidateSQL() should return an error for empty ident")
	}
}

func TestString_ValidateSQL_Empty(t *testing.T) {
	s := String("")

	err := s.ValidateSQL()
	if err != nil {
		t.Errorf("String.ValidateSQL() returned an error: %v", err)
	}
}

func TestInt_ValidateSQL_Empty(t *testing.T) {
	i := Int(0)

	err := i.ValidateSQL()
	if err != nil {
		t.Errorf("Int.ValidateSQL() returned an error: %v", err)
	}
}

func TestFloat_ValidateSQL_Empty(t *testing.T) {
	f := Float(0)

	err := f.ValidateSQL()
	if err != nil {
		t.Errorf("Float.ValidateSQL() returned an error: %v", err)
	}
}

func TestFloat_ValidateSQL(t *testing.T) {
	f := Float(3.14)
	err := f.ValidateSQL()
	if err != nil {
		t.Errorf("Float.ValidateSQL() returned an error: %v", err)
	}
}

type id int64

func TestInt64(t *testing.T) {
	var v any = id(1)
	_ = LValue(v)
}
