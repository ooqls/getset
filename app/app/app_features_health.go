package app

const (
	health_pathOpt string = "opt-health-path"
	health_intervalOpt string = "opt-health-interval"
)

var healthPathFlag string

type healthOpt struct {
	featureOpt
}

func WithHealthPath(path string) healthOpt {
	return healthOpt{
		featureOpt: featureOpt{
			key:   health_pathOpt,
			value: path,
		},
	}
}

func WithHealthInterval(interval int) healthOpt {
	return healthOpt{
		featureOpt: featureOpt{
			key:   health_intervalOpt,
			value: interval,
		},
	}
}

type HealthFeature struct {
	Enabled  bool
	Path     string
	Interval int
}

func (f *HealthFeature) apply(opt healthOpt) {
	switch opt.key {
	case health_pathOpt:
		f.Path = opt.value.(string)	
	case health_intervalOpt:
		f.Interval = opt.value.(int)
	}
}

func Health(opts ...healthOpt) HealthFeature {
	f := HealthFeature{
		Enabled:  true,
		Path:     healthPathFlag,
		Interval: 30,
	}
	for _, opt := range opts {
		f.apply(opt)
	}
	return f
}
