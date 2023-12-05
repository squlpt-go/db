package db

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

const (
	testParentId = 1
	testChildId1 = 1
	testChildId2 = 2
)

func TestFailInsertRow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	_, _ = InsertRow(DB(), &Child{})
}

func TestInsertDeleteRow(t *testing.T) {
	p := Parent{
		Name:   NewNullable("Name"),
		Status: active,
	}
	r, err := InsertRow(DB(), p)
	if err != nil {
		t.Error(err)
		return
	}

	h := Child{
		Name: "Child 123",
	}
	r, err = InsertRow(DB(), h)
	if err != nil {
		t.Error(err)
		return
	}

	if n, _ := r.RowsAffected(); n != 1 {
		t.Errorf("One row should have been inserted")
	}

	h.ID, _ = r.LastInsertId()
	r, err = DeleteRow(DB(), h)

	if err != nil {
		t.Error(err)
		return
	}

	if n, _ := r.RowsAffected(); n != 1 {
		t.Errorf("One row should have been deleted")
	}
}

func TestUpdateRow(t *testing.T) {
	existing, has := GetRows[Child](DB()).Row()
	if !has {
		t.Fatal("could not get existing child")
	}

	h := Child{
		ID:   existing.ID,
		Name: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	r, err := UpdateRow(DB(), h)

	if err != nil {
		t.Error(err)
		return
	}

	if n, _ := r.RowsAffected(); n != 1 {
		t.Errorf("One row should have been updated")
	}

	_, has = GetRowById[Child](DB(), existing.ID)

	if !has {
		t.Error("Child row doesn't exist")
	}
}

func TestColumns(t *testing.T) {
	db := DB()
	fields := GetTableFields(db, "parents")

	rows, err := db.Query("SELECT * FROM parents WHERE 0 = 1")
	if err != nil {
		t.Error(err)
		return
	}

	columns, err := rows.Columns()
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(fields, columns) {
		t.Errorf("%v (actual)\n%v (expected)", fields, columns)
	}
}

func TestCount(t *testing.T) {
	db := DB()
	row := db.QueryRow(`
		SELECT COUNT(*)
		FROM parents
		WHERE parent_status = 'active'`,
	)

	if row == nil {
		t.Error("invalid row")
		return
	}
	var count int64
	err := row.Scan(&count)
	if err != nil {
		t.Error(err)
		return
	}

	c := GetCount[Parent](
		db,
		NewQuery().WhereEq("parent_status", "active"),
	)

	if c != count {
		t.Error("Count mismatch")
	}
}

func TestGetRows(t *testing.T) {
	db := DB()
	r := GetRows[Child](db).Slice()

	if r[0].Parent.Name.Wrapped == "" {
		t.Error("Did not get Parent")
	}
}

func TestGetChildrenOneToMany(t *testing.T) {
	db := DB()
	existing, has := GetRows[Child](db).Row()
	if !has {
		t.Fatal("could not get child")
	}
	_ = GetChildren[Parent, Child](db, existing.ID)
}

func TestGetChildrenManyToMany(t *testing.T) {
	db := DB()
	existing, has := GetRows[Child](db).Row()
	if !has {
		t.Fatal("could not get child")
	}
	_ = GetChildren[Parent, Friend](db, existing.ID)
}

func TestAssignChildrenOneToMany(t *testing.T) {
	db := DB()

	AssignChildren[Parent, Child](db, testParentId, []int{}, true)
	c := GetChildren[Parent, Child](db, testParentId)
	if len(c.Slice()) != 0 {
		t.Error("Assignment failed")
	}

	AssignChildren[Parent, Child](db, testParentId, []int{testChildId1}, true)
	c = GetChildren[Parent, Child](db, testParentId)
	if len(c.Slice()) != 1 {
		t.Error("Assignment failed")
	}

	AssignChildren[Parent, Child](db, testParentId, []int{testChildId2}, false)
	c = GetChildren[Parent, Child](db, testParentId)
	if len(c.Slice()) != 2 {
		t.Error("Assignment failed")
	}
}

func TestSetChildrenOneToMany(t *testing.T) {
	db := DB()

	err := SetChildren[Parent, Child](db, testParentId, []Child{}, true)
	if err != nil {
		t.Error(err)
	}
	c := GetChildren[Parent, Child](db, testParentId)
	if len(c.Slice()) != 0 {
		t.Error("Assignment failed")
	}

	err = SetChildren[Parent, Child](db, testParentId, []Child{{Name: "New Child 1"}}, true)
	if err != nil {
		t.Error(err)
	}
	cs := GetChildren[Parent, Child](db, testParentId).Slice()
	if len(cs) != 1 {
		t.Error("Assignment failed")
	}

	err = SetChildren[Parent, Child](db, testParentId, []Child{{Name: "New Child 2"}}, false)
	if err != nil {
		t.Error(err)
	}
	cs = GetChildren[Parent, Child](db, testParentId).Slice()
	if len(cs) != 2 {
		t.Error("Assignment failed")
	}
}

func TestSetChildrenManyToMany(t *testing.T) {
	db := DB()

	err := SetChildren[Parent, Friend](db, testParentId,
		[]Friend{
			{Name: "test1"},
		}, true)
	if err != nil {
		t.Error(err)
	}

	c := GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 1 {
		t.Error("Assignment failed")
	}

	err = SetChildren[Parent, Friend](db, testParentId,
		[]Friend{
			{Name: "test2"},
		}, true)
	if err != nil {
		t.Error(err)
	}

	c = GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 1 {
		t.Error("Assignment failed")
	}

	err = SetChildren[Parent, Friend](db, testParentId,
		[]Friend{
			{ID: 600, Name: "test3"},
		}, false)
	if err != nil {
		t.Error(err)
	}

	c = GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 2 {
		t.Error("Assignment failed")
	}

	err = SetChildren[Parent, Friend](db, testParentId,
		[]Friend{
			{ID: 600, Name: "test4"},
			{Name: "test5"},
			{Name: "test6"},
		}, false)
	if err != nil {
		t.Error(err)
	}

	c = GetChildren[Parent, Friend](db, testParentId)
	s := c.Slice()
	if len(s) != 4 {
		t.Error("Assignment failed", s)
	}

	err = SetChildren[Parent, Friend](db, testParentId,
		[]Friend{
			{Name: "test7"},
		}, true)
	if err != nil {
		t.Error(err)
	}

	c = GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 1 {
		t.Error("Assignment failed")
	}
}

func TestAssignChildrenManyToMany(t *testing.T) {
	db := DB()

	AssignChildren[Parent, Friend](db, testParentId, []int{testChildId1}, true)

	c := GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 1 {
		t.Error("Assignment failed")
	}

	AssignChildren[Parent, Friend](db, testParentId, []int{testChildId1}, true)

	c = GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 1 {
		t.Error("Assignment failed")
	}

	AssignChildren[Parent, Friend](db, testParentId, []int{testChildId2}, false)

	c = GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 2 {
		t.Error("Assignment failed")
	}
	AssignChildren[Parent, Friend](db, testParentId, []int{testChildId2}, true)

	c = GetChildren[Parent, Friend](db, testParentId)
	if len(c.Slice()) != 1 {
		t.Error("Assignment failed")
	}
}
