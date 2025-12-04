package store_test

import (
	"cronlogger/store"
	"database/sql"
	"fmt"
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

func Test_Paged_Results(t *testing.T) {
	s, db := getStore(t)
	defer db.Close()

	for i := range 10 {
		s.Create(store.OpResultEntity{
			App:     fmt.Sprintf("test_%d", i),
			Success: true,
			Output:  "",
		})
	}

	// standards queries
	// ----------------------------------------------------------------------

	// retrieve all 10 items
	res, err := s.GetPagedItems(10, 0, nil, nil)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 10 {
		t.Errorf("expected 10 total items, got %d", res.TotalCount)
	}

	// retrieve 5 items
	res, err = s.GetPagedItems(5, 0, nil, nil)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 10 {
		t.Errorf("expected 10 total items, got %d", res.TotalCount)
	}
	if len(res.Items) != 5 {
		t.Errorf("expected 5 items, got %d", res.TotalCount)
	}
	if res.Items[0].App != "test_9" {
		t.Errorf("expected test_9 item, got %s", res.Items[0].App)
	}
	if res.Items[4].App != "test_5" {
		t.Errorf("expected test_5 item, got %s", res.Items[0].App)
	}

	// retrieve 3 items / skip 3
	res, err = s.GetPagedItems(3, 3, nil, nil)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 10 {
		t.Errorf("expected 10 total items, got %d", res.TotalCount)
	}
	if len(res.Items) != 3 {
		t.Errorf("expected 3 items, got %d", res.TotalCount)
	}
	if res.Items[0].App != "test_6" {
		t.Errorf("expected test_6 item, got %s", res.Items[0].App)
	}
	if res.Items[2].App != "test_4" {
		t.Errorf("expected test_4 item, got %s", res.Items[0].App)
	}

	// filter date
	future := time.Now().AddDate(1, 0, 0)
	res, err = s.GetPagedItems(10, 0, &future, nil)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 0 {
		t.Errorf("expected 0 total items, got %d", res.TotalCount)
	}

	past := time.Now().AddDate(-1, 0, 0)
	res, err = s.GetPagedItems(10, 0, nil, &past)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 0 {
		t.Errorf("expected 0 total items, got %d", res.TotalCount)
	}

	res, err = s.GetPagedItems(10, 0, &past, &future)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 10 {
		t.Errorf("expected 10 total items, got %d", res.TotalCount)
	}

	// corner cases
	// ----------------------------------------------------------------------

	res, err = s.GetPagedItems(0, 0, nil, nil)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 10 {
		t.Errorf("expected 10 total items, got %d", res.TotalCount)
	}
	if len(res.Items) != 0 {
		t.Errorf("expected 0 items, got %d", res.TotalCount)
	}

	// negative pagesize
	res, err = s.GetPagedItems(-2, 0, nil, nil)
	if err == nil {
		t.Errorf("error expected for negative pagesize")
	}

	// negative skip
	res, err = s.GetPagedItems(0, -3, nil, nil)
	if err == nil {
		t.Errorf("error expected for negative offset")
	}

	// skip is too big
	res, err = s.GetPagedItems(0, 100, nil, nil)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 10 {
		t.Errorf("expected 10 total items, got %d", res.TotalCount)
	}
	if len(res.Items) != 0 {
		t.Errorf("expected 0 items, got %d", res.TotalCount)
	}

	// retrieve 3 items / skip 8
	res, err = s.GetPagedItems(3, 8, nil, nil)
	if err != nil {
		t.Errorf("could not get paged items; %v", err)
	}
	if res.TotalCount != 10 {
		t.Errorf("expected 10 total items, got %d", res.TotalCount)
	}
	if len(res.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(res.Items))
	}
	if res.Items[0].App != "test_1" {
		t.Errorf("expected test_1 item, got %s", res.Items[0].App)
	}
	if res.Items[1].App != "test_0" {
		t.Errorf("expected test_0 item, got %s", res.Items[1].App)
	}

}
