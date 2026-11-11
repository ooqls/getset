package postgres

import (
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config returns explicit Postgres driver settings for SQLX and PGX. Registry
// credentials and TLS must not be replaced by environment or passfile defaults.
// Callers own the returned snapshot; TLS has no plaintext fallback.
func (opt Options) Config() (*pgxpool.Config, error) {
	if opt.Host == "" || strings.ContainsAny(opt.Host, ",\x00") {
		return nil, fmt.Errorf("Postgres requires a single host")
	}
	if opt.Port < 1 || opt.Port > 65535 {
		return nil, fmt.Errorf("invalid Postgres port %d", opt.Port)
	}
	// pgx reads PGSERVICE before applying the resulting configuration. Reject
	// this ambient override rather than read an arbitrary service file.
	if os.Getenv("PGSERVICE") != "" {
		return nil, fmt.Errorf("PGSERVICE is not supported: configure Postgres in the Registry")
	}
	// Parse a safe baseline to satisfy pgx's construction contract. Disable
	// TLS/file lookup before parsing, then assign resolved settings directly.
	baseline := "host=localhost port=5432 user='' password='' dbname=postgres " +
		"sslmode=disable sslrootcert='' sslcert='' sslkey='' passfile='" + os.DevNull + "' " +
		"connect_timeout=0 target_session_attrs=any sslnegotiation=postgres " +
		"min_protocol_version=3.0 max_protocol_version=3.0 channel_binding=prefer require_auth=''"
	cfg, err := pgxpool.ParseConfig(baseline)
	if err != nil {
		return nil, fmt.Errorf("configure Postgres driver: %w", err)
	}
	conn := cfg.ConnConfig
	conn.Host, conn.Port = opt.Host, uint16(opt.Port)
	conn.User, conn.Password, conn.Database = opt.User, opt.Pw, opt.DB
	conn.RuntimeParams = map[string]string{}
	conn.Fallbacks = nil
	if opt.Tls != nil {
		if strings.HasPrefix(opt.Host, "/") || strings.HasPrefix(opt.Host, "@") {
			return nil, fmt.Errorf("TLS requires a TCP Postgres host")
		}
		conn.TLSConfig = opt.Tls.Clone()
		if conn.TLSConfig.ServerName == "" {
			conn.TLSConfig.ServerName = opt.Host
		}
	} else {
		conn.TLSConfig = nil
	}
	return cfg, nil
}
