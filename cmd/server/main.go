package main

import (
	"context"
	"cronlogger"
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
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/viper"
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
		port       int
		host       string
		dbPath     string
		logLevel   string
		configFile string
		help       bool
	)

	flag.IntVar(&port, "port", 9000, "define the port of the server")
	flag.StringVar(&host, "host", "localhost", "define the hostname of the server")
	flag.StringVar(&dbPath, "db", "./cronlog-store.db", "the path to the db file")
	flag.StringVar(&logLevel, "loglevel", "INFO", "the loglevel to use (DEBUG|INFO|WARN|ERROR)")
	flag.StringVar(&configFile, "config", "./", "the path to the config file")
	flag.BoolVar(&help, "help", false, "show the help information")
	flag.Parse()

	ver := fmt.Sprintf("%s-%s", Version, Build)
	if help {
		fmt.Printf("Cronlogger %s\n\n", ver)
		flag.Usage()
		os.Exit(0)
	}

	config, err := readConfig(configFile)
	if err != nil {
		fmt.Printf("%v, exiting", err)
		os.Exit(1)
	}

	store, db, err := store.CreateSqliteStoreFromDbPath(dbPath)
	if err != nil {
		fmt.Printf("%v, exiting", err)
		os.Exit(1)
	}
	defer db.Close()

	handler := handler.New(store, setupLogging(logLevel), ver, config)
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

func readConfig(configPath string) (cronlogger.AppConfig, error) {
	viper.SetConfigName("application")
	viper.SetConfigType("yaml")

	// Add search paths to find the file
	viper.AddConfigPath("/etc/cronlogger/")
	viper.AddConfigPath("/var/cronlogger/")
	viper.AddConfigPath("$HOME/.cronlogger")

	configPath = filepath.Clean(configPath)
	configPath, err := filepath.Abs(configPath)
	if err == nil {
		_, err = os.Stat(configPath)
		if !os.IsNotExist(err) {
			viper.AddConfigPath(configPath)
		}
	} else {
		fmt.Printf("cannot resolve supplied path '%s'; %v", configPath, err)
	}

	// current executable path
	path, err := os.Executable()
	if err == nil {
		basePath := filepath.Dir(path)
		viper.AddConfigPath(basePath)
	}

	var config cronlogger.AppConfig
	err = viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("cannot read configuration; %v", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("cannot parse configuration; %v", err)
	}
	return config, nil

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
