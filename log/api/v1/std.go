package v1

import (
	"encoding/json"
	"net/http"

	"github.com/ooqls/getset/log"
	"github.com/ooqls/getset/log/api/std/v1/gen"
	schemas "github.com/ooqls/getset/log/api/v1/gen"
)

var _ gen.ServerInterface = &LoggingServerStdImpl{}

type LoggingServerStdImpl struct{}

func (l *LoggingServerStdImpl) GetLoggingConfig(w http.ResponseWriter, r *http.Request) {
	cfg := schemas.LoggingConfig{
		Format: log.GetFormat(),
		Level:  log.GetLevel(),
	}
	err := json.NewEncoder(w).Encode(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (l *LoggingServerStdImpl) SetLoggingConfig(w http.ResponseWriter, r *http.Request) {
	cfg := schemas.LoggingConfig{}
	err := json.NewDecoder(r.Body).Decode(&cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.SetFormat(cfg.Format)
	log.SetLogLevel(cfg.Level)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (l *LoggingServerStdImpl) UpdateLoggingLevel(w http.ResponseWriter, r *http.Request, level schemas.LevelEnum) {
	log.SetLogLevel(level)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (l *LoggingServerStdImpl) UpdateLoggingFormat(w http.ResponseWriter, r *http.Request, format schemas.FormatEnum) {
	log.SetFormat(format)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
