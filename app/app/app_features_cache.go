package app

type cacheOpt struct {
	featureOpt
}

const (
	cache_typeOpt string = "opt-cache-type"
)

type cacheType string

const (
	cacheTypeRedis  cacheType = "redis"
	cacheTypeMem    cacheType = "mem"
	cacheTypeValkey cacheType = "valkey"
)

func WithCacheType(cacheType string) cacheOpt {
	return cacheOpt{featureOpt: featureOpt{key: cache_typeOpt, value: cacheType}}
}

type CacheFeature struct {
	Enabled   bool
	CacheType cacheType
}

func (f *CacheFeature) apply(opt cacheOpt) {
	switch opt.key {
	case cache_typeOpt:
		f.CacheType = opt.value.(cacheType)
	}
}

func Cache(opts ...cacheOpt) CacheFeature {
	f := CacheFeature{
		Enabled:   true,
		CacheType: cacheTypeRedis,
	}
	for _, opt := range opts {
		f.apply(opt)
	}
	return f
}
