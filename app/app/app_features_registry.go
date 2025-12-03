package app

const (
	registry_pathOpt string = "opt-registry-path"
)

var registryPathFlag string

type registryOpt struct {
	featureOpt
}

func WithRegistryPath(p string) registryOpt {
	return registryOpt{featureOpt: featureOpt{
			key: registry_pathOpt,
			value: &p,
		},
	}
}

type RegistryFeature struct {
	enabled      bool
	registryPath *string
}

// GetRegistryPath returns the registry path or empty string if nil
func (f *RegistryFeature) RegistryPath() string {
	if f.registryPath == nil {
		return ""
	}
	return *f.registryPath
}

func (f *RegistryFeature) apply(opt registryOpt) {
	switch opt.key {
	case registry_pathOpt:
		f.registryPath = opt.value.(*string)
	}
}

func Registry(opts ...registryOpt) RegistryFeature {
	f := RegistryFeature{
		enabled: true,
		registryPath: &registryPathFlag,
	}

	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}

