package app

import (
	"net/http"
)

const (
	http_portOpt string = "opt-http-port"
	http_muxOpt  string = "opt-http-mux"
)

type httpOpt struct {
	featureOpt
}

func WithHttpPort(port int) httpOpt {
	return httpOpt{
		featureOpt: featureOpt{
			key:   http_portOpt,
			value: port,
		},
	}
}

func WithHttpMux(mux *http.ServeMux) httpOpt {
	return httpOpt{
		featureOpt: featureOpt{
			key:   http_muxOpt,
			value: mux,
		},
	}
}

type HTTPFeature struct {
	Enabled bool
	Port    int
	Mux     *http.ServeMux
}

func HTTP(opts ...httpOpt) HTTPFeature {
	f := HTTPFeature{
		Enabled: true,
		Mux:     http.NewServeMux(),
	}

	for _, o := range opts {
		switch o.key {
		case http_portOpt:
			f.Port = o.value.(int)
		case http_muxOpt:
			f.Mux = o.value.(*http.ServeMux)
		}
	}

	return f
}
