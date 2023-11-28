package db

import (
	"database/sql"
	"testing"
)

func init() {
	db := DB()

	ManyToOne[Child, Parent](db)
	OneToMany[Parent, Child](db)
	ManyToMany[Parent, Friend](db, "parent_friends")
}

func TestManyToOne(t *testing.T) {
	db := DB()
	expectedRelation := ManyToOneDef{
		"children",
		"children.parent_id",
		"parents",
		"parents.parent_id",
	}
	assertRelationExists(t, db, expectedRelation)
}

func TestOneToMany(t *testing.T) {
	db := DB()
	expectedRelation := OneToManyDef{
		"parents",
		"parents.parent_id",
		"children",
		"children.parent_id",
	}
	assertRelationExists(t, db, expectedRelation)
}

func TestManyToMany(t *testing.T) {
	db := DB()
	expectedRelation := ManyToManyDef{
		"parents",
		"parents.parent_id",
		"friends",
		"friends.friend_id",
		"parent_friends",
		"parent_friends.parent_id",
		"parent_friends.friend_id",
	}
	assertRelationExists(t, db, expectedRelation)
}

func TestDefRelation(t *testing.T) {
	db := DB()
	expectedRelation := ManyToOneDef{
		"children1",
		"children1.parent_id",
		"parents1",
		"parents1.parent_id",
	}
	DefRelation(db, expectedRelation)
	assertRelationExists(t, db, expectedRelation)
}

func assertRelationExists(t *testing.T, db *sql.DB, expectedRelation Relation) {
	relations, ok := schema[db]
	if !ok {
		t.Errorf("Expected relation not found in schema for DB.")
		return
	}

	found := false
	for _, relation := range relations {
		if relation == expectedRelation {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected relation not found in schema for DB.")
	}
}
