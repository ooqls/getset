package app

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	gin_engineOpt string = "opt-gin-engine"
	gin_portOpt   string = "opt-gin-port"
	gin_corsOpt   string = "opt-gin-cors"
)

type ginOpt struct {
	featureOpt
}

func WithGin(e *gin.Engine) ginOpt {
	return ginOpt{
		featureOpt: featureOpt{
			key:   gin_engineOpt,
			value: e,
		},
	}
}

func WithGinCors(cors cors.Config) ginOpt {
	return ginOpt{
		featureOpt: featureOpt{
			key:   gin_corsOpt,
			value: &cors,
		},
	}
}

type GinFeature struct {
	Enabled bool
	Port    int
	Cors    *cors.Config
	Engine  *gin.Engine
}

func (f *GinFeature) apply(opt ginOpt) {
	switch opt.key {
	case gin_engineOpt:
		f.Engine = opt.value.(*gin.Engine)
	case gin_portOpt:
		f.Port = opt.value.(int)
	case gin_corsOpt:
		f.Cors = opt.value.(*cors.Config)
	}
}

func Gin(opts ...ginOpt) GinFeature {
	f := GinFeature{
		Enabled: true,
		Port:    8080,
		Engine:  gin.New(),
	}

	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}
