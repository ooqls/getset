package app

const (
	logging_api_portOpt string = "opt-logging-api-port"
)

var loggingApiPortFlag int

type loggingApiOpt struct {
	featureOpt
}

func WithLoggingApiPort(port int) loggingApiOpt {
	return loggingApiOpt{featureOpt: featureOpt{key: logging_api_portOpt, value: port}}
}

type LoggingApiFeature struct {
	Enabled bool
	Port    int
}

func (f *LoggingApiFeature) apply(opt loggingApiOpt) {
	switch opt.key {
	case logging_api_portOpt:
		f.Port = opt.value.(int)
	}
}

func LoggingApi(opts ...loggingApiOpt) LoggingApiFeature {
	f := LoggingApiFeature{
		Enabled: true,
		Port: loggingApiPortFlag,
	}
	
	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}