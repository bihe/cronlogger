package store

import (
	"context"
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
	GetById(id string) (OpResultEntity, error)
	GetAll() ([]OpResultEntity, error)
	GetPagedItems(pageSize, skip int, from, until *time.Time, appName string) (PagedOpResults, error)
	GetAvailApps() ([]string, error)
}

// CreateStore creates a new store to persist data
func CreateStore(con Connection) OpResultStore {
	return &dbStore{
		con: con,
	}
}

// CreateSqliteStoreFromDbPath initializes a new store from a sqlite file path
func CreateSqliteStoreFromDbPath(dbPath string) (OpResultStore, *sql.DB, error) {
	db := MustCreateSqliteConn(dbPath)
	con, err := CreateGormSqliteCon(db)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create database connection: %v", err)
	}

	// Migrate the schema
	con.Write.AutoMigrate(&OpResultEntity{})
	con.Read.AutoMigrate(&OpResultEntity{})

	return CreateStore(con), db, nil
}

type PagedOpResults struct {
	TotalCount int64
	Items      []OpResultEntity
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbStore struct {
	con Connection
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

func (s *dbStore) GetById(id string) (OpResultEntity, error) {
	if id == "" {
		return OpResultEntity{}, fmt.Errorf("no id supplied")
	}

	ctx := context.Background()
	item, err := gorm.G[OpResultEntity](s.con.R()).Where("id = ?", id).First(ctx)
	if err != nil {
		return OpResultEntity{}, fmt.Errorf("could not retrieve all entries; %v", err)
	}
	return item, nil
}

func (s *dbStore) GetAll() ([]OpResultEntity, error) {
	ctx := context.Background()
	results, err := gorm.G[OpResultEntity](s.con.R()).Order("created DESC").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve all entries; %v", err)
	}
	return results, nil
}

func (s *dbStore) GetPagedItems(pageSize, skip int, from, until *time.Time, appName string) (PagedOpResults, error) {
	if pageSize < 0 {
		return PagedOpResults{}, fmt.Errorf("negative pagesizes do not make sense")
	}

	if skip < 0 {
		return PagedOpResults{}, fmt.Errorf("negative offset does not make sense")
	}

	where := ""
	params := make([]any, 0)
	if from != nil && until != nil {
		where = "created >= ? and created <= ?"
		params = append(params, *from, *until)
	} else if from != nil {
		where = "created >= ?"
		params = append(params, *from)
	} else if until != nil {
		where = "created <= ?"
		params = append(params, *until)
	}

	if appName != "" {
		if where != "" {
			where += " and "
		}
		where += "application = ?"
		params = append(params, appName)
	}

	var (
		results      []OpResultEntity
		totalEntries int64
	)

	query := s.con.R().Find(&results).Order("created DESC")
	if where != "" {
		query = s.con.R().Where(where, params...).Find(&results).Order("created DESC")
	}

	g := query.Count(&totalEntries)
	if g.Error != nil {
		return PagedOpResults{}, fmt.Errorf("could not retrieve count of entries; %v", g.Error)
	}
	g = query.Limit(pageSize).Offset(skip).Find(&results)
	if g.Error != nil {
		return PagedOpResults{}, fmt.Errorf("could not retrieve entries; %v", g.Error)
	}

	return PagedOpResults{Items: results, TotalCount: totalEntries}, nil
}

func (s *dbStore) GetAvailApps() ([]string, error) {
	var apps []string
	g := s.con.R().Model(&OpResultEntity{}).Group("application").Order("application ASC").Pluck("application", &apps)
	if g.Error != nil {
		return nil, g.Error
	}
	return apps, nil
}
