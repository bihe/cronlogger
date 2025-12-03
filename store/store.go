package store

import (
	"context"
	"cronlogger/persistence"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// the store defines a very simple interface to store the result of an operation
// (typically a shell-script) in a table.

// An OpResultEntity the result of an execution
type OpResultEntity struct {
	ID      string    `gorm:"primary_key;TYPE:varchar(36);COLUMN:id"`
	App     string    `gorm:"COLUMN:application;TYPE:nvarchar(255);"`
	Success bool      `gorm:"COLUMN:success;TYPE:bool;DEFAULT:FALSE;NOT NULL"`
	Output  string    `gorm:"COLUMN:output;TYPE:nvarchar(255);"`
	Created time.Time `gorm:"COLUMN:created;NOT NULL"`
}

// TableName specifies the name of the Table used
func (OpResultEntity) TableName() string {
	return "OPRESULTS"
}

// OpResultStore provides methods to interact with the store
type OpResultStore interface {
	Create(item OpResultEntity) (OpResultEntity, error)
	GetAll() ([]OpResultEntity, error)
}

// CreateStore creates a new store to persist data
func CreateStore(con persistence.Connection) OpResultStore {
	return &dbStore{
		con: con,
	}
}

// CreateSqliteStoreFromDbPath initializes a new store from a sqlite file path
func CreateSqliteStoreFromDbPath(dbPath string) (OpResultStore, *sql.DB, error) {
	db := persistence.MustCreateSqliteConn(dbPath)
	con, err := persistence.CreateGormSqliteCon(db)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create database connection: %v", err)
	}

	// Migrate the schema
	con.Write.AutoMigrate(&OpResultEntity{})
	con.Read.AutoMigrate(&OpResultEntity{})

	return CreateStore(con), db, nil
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbStore struct {
	con persistence.Connection
}

func (s *dbStore) Create(item OpResultEntity) (OpResultEntity, error) {
	// set the necessary values like a new ID and created date
	item.ID = uuid.New().String()
	item.Created = time.Now()
	ctx := context.Background()
	err := gorm.G[OpResultEntity](s.con.W()).Create(ctx, &item)
	if err != nil {
		return OpResultEntity{}, fmt.Errorf("could not store a new item: %v", err)
	}
	return item, nil
}

func (s *dbStore) GetAll() ([]OpResultEntity, error) {
	ctx := context.Background()
	results, err := gorm.G[OpResultEntity](s.con.W()).Order("created DESC").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve all entries; %v", err)
	}
	return results, nil
}
