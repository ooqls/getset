package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gingen "github.com/ooqls/getset/log/api/gin/v1/gen"
	stdgen "github.com/ooqls/getset/log/api/std/v1/gen"
)

func Gin(e *gin.Engine) {
	gingen.RegisterHandlers(e, &LoggingServerGinImpl{})
}

func Std() http.Handler {
	return stdgen.Handler(&LoggingServerStdImpl{})
}
