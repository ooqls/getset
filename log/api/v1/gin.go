package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/getset/log/api/v1/gen"
)

type LoggingServerGinImpl struct{}

func (l *LoggingServerGinImpl) GetLoggingConfig(c *gin.Context) {
	cfg := gen.LoggingConfig{
		Format: log.GetFormat(),
		Level:  log.GetLevel(),
	}

	// Implement the logic to get the logging configuration
	// For example, you might want to read from a config file or environment variable
	c.JSON(200, cfg)
}

func (l *LoggingServerGinImpl) SetLoggingConfig(c *gin.Context, cfg gen.LoggingConfig) {
	// Implement the logic to set the logging configuration
	// For example, you might want to write to a config file or environment variable
	log.SetFormat(cfg.Format)
	log.SetLogLevel(cfg.Level)

	c.JSON(200, gin.H{
		"message": "Logging configuration updated successfully",
	})
}

func (l *LoggingServerGinImpl) UpdateLoggingLevel(c *gin.Context, level gen.LevelEnum) {
	// Implement the logic to set the log level
	// For example, you might want to write to a config file or environment variable
	log.SetLogLevel(level)

	c.JSON(200, level)
}

func (l *LoggingServerGinImpl) UpdateLoggingFormat(c *gin.Context, f gen.FormatEnum) {
	log.SetFormat(gen.FormatEnum(f))
	c.JSON(200, gin.H{
		"message": "Logging format updated successfully",
	})

}
