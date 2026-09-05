package db

import (
	"embed"
	"fmt"

	"github.com/TGScheme/TLExtractorBot/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/valkey-io/valkey-go"
)

//go:embed schema/*.sql
var schemaFiles embed.FS

func NewDB(cfg *config.Config) (*DB, error) {
	dsn := cfg.DSN()
	if err := goose.SetDialect("pgx"); err != nil {
		return nil, fmt.Errorf("goose dialect: %w", err)
	}
	sqlConn, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("goose open: %w", err)
	}
	if err = waitFor("postgres", sqlConn.Ping); err != nil {
		_ = sqlConn.Close()
		return nil, err
	}
	goose.SetBaseFS(schemaFiles)
	goose.SetLogger(goose.NopLogger())
	if err = goose.Up(sqlConn, "schema"); err != nil {
		_ = sqlConn.Close()
		return nil, fmt.Errorf("goose up: %w", err)
	}
	_ = sqlConn.Close()

	var redis valkey.Client
	if err = waitFor("valkey", func() error {
		redis, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{cfg.ValkeyAddr},
		})
		return err
	}); err != nil {
		return nil, err
	}
	return new(dsn, redis)
}
