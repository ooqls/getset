package app

type AppState struct {
	RegistryInitialized   bool
	JWTInitialized        bool
	RSAInitialized        bool
	LoggingAPIInitialized bool
	CacheInitialized      bool
	HTTPInitialized       bool
	GinInitialized        bool
	GrpcInitialized       bool
	SQLiteInitialized     bool
	DocsInitialized       bool
	TLSInitialized        bool
	SQLInitialized        bool
	RedisInitialized      bool
	ValkeyInitialized     bool
	SQLSeeded             bool
	Healthy               bool
	Running               bool
}
