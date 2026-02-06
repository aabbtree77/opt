package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"app.root/config"
	dbpkg "app.root/db"
	"app.root/routes"
)

func main() {
	fmt.Println("app.root starting")

	// -----------------------------------------------------
	// Load config
	// -----------------------------------------------------
	cfg := config.LoadConfig()

	fmt.Println("Configuration loaded.")

	// -----------------------------------------------------
	// Database (retry loop)
	// -----------------------------------------------------
	var db *sql.DB
	var err error

	for i := 1; i <= 15; i++ {
		db, err = sql.Open("pgx", cfg.DBDSN)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = db.PingContext(ctx)
			cancel()
		}

		if err == nil {
			break
		}

		fmt.Printf("database not ready (attempt %d/15): %v\n", i, err)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		panic(fmt.Errorf("database connection failed after retries: %w", err))
	}

	defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	fmt.Println("database ready")

	// -----------------------------------------------------
	// Migrations
	// -----------------------------------------------------
	if err := dbpkg.RunMigrations(db); err != nil {
		panic(err)
	}

	fmt.Println("migrations OK")

	// -----------------------------------------------------
	// HTTP server
	// -----------------------------------------------------
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, db, &cfg)

	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		fmt.Println("listening on", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// -----------------------------------------------------
	// Graceful shutdown
	// -----------------------------------------------------
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-sigCtx.Done()
	fmt.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)

	fmt.Println("bye")
}
