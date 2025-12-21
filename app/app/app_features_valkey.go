package app

import "github.com/ooqls/getset/registry"

const (
	valkey_hostOpt     string = "opt-valkey-host"
	valkey_UserOpt     string = "opt-valkey-user"
	valkey_passwordOpt string = "opt-valkey-password"
	valkey_port        string = "opt-valkey-port"
)

func WithValkeyHost(host string) valkeyOpt {
	return valkeyOpt{featureOpt: featureOpt{key: valkey_hostOpt, value: host}}
}

func WithValkeyUser(username string) valkeyOpt {
	return valkeyOpt{featureOpt: featureOpt{key: valkey_UserOpt, value: username}}
}

func WithValkeyPassword(password string) valkeyOpt {
	return valkeyOpt{featureOpt: featureOpt{key: valkey_passwordOpt, value: password}}
}

func WithValkeyPort(port int) valkeyOpt {
	return valkeyOpt{featureOpt: featureOpt{key: valkey_port, value: port}}
}

type ValkeyFeature struct {
	Enabled  bool
	valkeyDB registry.Database
}

func (f *ValkeyFeature) apply(opt valkeyOpt) {
	switch opt.key {
	case valkey_hostOpt:
		f.valkeyDB.Server.Host = opt.value.(string)
	case valkey_UserOpt:
		f.valkeyDB.Auth.Username = opt.value.(string)
	case valkey_passwordOpt:
		f.valkeyDB.Auth.Password = opt.value.(string)
	case valkey_port:
		f.valkeyDB.Server.Port = opt.value.(int)
	}
}

type valkeyOpt struct {
	featureOpt
}

func Valkey(opts ...valkeyOpt) ValkeyFeature {
	r := registry.Get()	

	valkeyDb := registry.Database{}
	if r.Valkey != nil {
		valkeyDb = *r.Valkey
	}

	f := ValkeyFeature{
		Enabled:  true,
		valkeyDB: valkeyDb,
	}

	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}
