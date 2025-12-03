package main

import (
	"context"
	"cronlogger/handler"
	"cronlogger/store"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "cronloggerserver"
)

// start a http server to show the result of the collected data of the cronlogger
func main() {
	var (
		port     int
		host     string
		dbPath   string
		logLevel string
		help     bool
	)

	flag.IntVar(&port, "port", 9000, "define the port of the server")
	flag.StringVar(&host, "host", "localhost", "define the hostname of the server")
	flag.StringVar(&dbPath, "db", "./cronlog-store.db", "the path to the db file")
	flag.StringVar(&logLevel, "loglevel", "INFO", "the loglevel to use (DEBUG|INFO|WARN|ERROR)")
	flag.BoolVar(&help, "help", false, "show the help information")
	flag.Parse()

	if help {
		fmt.Printf("Cronlogger %s-%s\n\n", Version, Build)
		flag.Usage()
		os.Exit(0)
	}

	store, db, err := store.CreateSqliteStoreFromDbPath(dbPath)
	if err != nil {
		panic(fmt.Sprintf("%v, exiting", err))
	}
	defer db.Close()

	handler := handler.New(store, setupLogging(logLevel))
	startServer(fmt.Sprintf("%s:%d", host, port), handler)
}

func setupLogging(level string) *slog.Logger {
	logLevel := &slog.LevelVar{} // INFO

	switch level {
	case "INFO":
		logLevel.Set(slog.LevelInfo)
	case "DEBUG":
		logLevel.Set(slog.LevelInfo)
	case "WARN":
		logLevel.Set(slog.LevelWarn)
	case "ERROR":
		logLevel.Set(slog.LevelError)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	return logger
}

func printServerBanner(name, version, build, addr string) {
	fmt.Printf("%s Starting server '%s'\n", "🚀", name)
	fmt.Printf("%s Version: '%s-%s'\n", "🔖", version, build)
	fmt.Printf("%s Listening on '%s'\n", "💻", addr)
	fmt.Printf("%s Ready!\n", "🏁")
}

func startServer(addr string, hdlr *handler.CronLogHandler) {
	mux := http.NewServeMux()
	handler.SetupRoutes(mux, hdlr)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	printServerBanner(AppName, Version, Build, addr)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
}
