package main

import (
	"cronlogger"
	"cronlogger/persistence"
	"database/sql"
	"flag"
	"fmt"
	"os"
)

// reads the stdin passed on via a pipe
// the result-code is passed via a shell variable - this is typically $?
func main() {
	var (
		exitCode int
		appName  string
		dbPath   string
	)
	flag.IntVar(&exitCode, "code", -1, "the exit-code of the command")
	flag.StringVar(&appName, "app", "", "the name of the application")
	flag.StringVar(&dbPath, "db", "", "the path to the db file")
	flag.Parse()

	if len(os.Args[1:]) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	if appName == "" {
		fmt.Println("No application-name supplied, exiting!")
		os.Exit(1)
	}

	result, err := cronlogger.ReadStdin()
	if err != nil {
		fmt.Printf("Could not read from Stdin: %v, exiting!\n", err)
		os.Exit(1)
	}
	if result != "" {
		// got the piped result of the command, store the result
		db := persistence.MustCreateSqliteConn(dbPath)
		defer db.Close()
		store := createStore(db)

		var success bool
		if exitCode == 0 {
			success = true
		}

		_, err := store.Create(cronlogger.OpResultEntity{
			App:     appName,
			Success: success,
			Output:  result,
		})

		if err != nil {
			fmt.Printf("Could not save item to store: %v, exiting!\n", err)
			os.Exit(1)
		}
	}
}

func createStore(db *sql.DB) cronlogger.OpResultStore {
	con, err := persistence.CreateGormSqliteCon(db)
	if err != nil {
		fmt.Printf("Could not create database connection: %v, exiting!\n", err)
		os.Exit(1)
	}

	// Migrate the schema
	con.Write.AutoMigrate(&cronlogger.OpResultEntity{})
	con.Read.AutoMigrate(&cronlogger.OpResultEntity{})

	return cronlogger.CreateStore(con)
}
