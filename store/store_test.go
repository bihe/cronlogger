package store_test

import (
	"cronlogger/store"
	"database/sql"
	"testing"
	"time"
)

func getStore(t *testing.T) (store.OpResultStore, *sql.DB) {
	store, db, err := store.CreateSqliteStoreFromDbPath(":memory:")
	if err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	return store, db
}

func Test_Create_OpResult(t *testing.T) {
	s, db := getStore(t)
	defer db.Close()

	items, err := s.GetAll()
	if err != nil {
		t.Errorf("could not get all items; %v", err)
	}
	if len(items) != 0 {
		t.Error("there should be not items in the result")
	}

	item, err := s.Create(store.OpResultEntity{
		App:     "test",
		Success: true,
		Output:  "",
	})
	if err != nil {
		t.Errorf("could not create an item; %v", err)
	}

	if item.ID == "" {
		t.Errorf("a created item must have an ID")
	}
	date := time.Date(2025, time.December, 1, 00, 0, 0, 0, time.UTC)
	if !item.Created.After(date) {
		t.Errorf("item needs to have a created timestamp")
	}

	items, err = s.GetAll()
	if err != nil {
		t.Errorf("could not get all items; %v", err)
	}

	if len(items) == 0 {
		t.Error("expected a list of items but got empty list")
	}
}

func Test_Result_Order(t *testing.T) {
	s, db := getStore(t)
	defer db.Close()

	s.Create(store.OpResultEntity{
		App:     "test1",
		Success: true,
		Output:  "",
	})
	s.Create(store.OpResultEntity{
		App:     "test2",
		Success: false,
		Output:  "",
	})
	s.Create(store.OpResultEntity{
		App:     "test3",
		Success: false,
		Output:  "",
	})

	items, err := s.GetAll()
	if err != nil {
		t.Errorf("could not get all items; %v", err)
	}
	if items[0].App != "test3" && items[2].App != "test1" {
		t.Errorf("the ordering of the results does not work")
	}
}
