package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	"github.com/ooqls/getset/db/redis"
	"github.com/ooqls/getset/db/valkey"
	v1 "github.com/ooqls/getset/log/api/v1"
	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

func (a *App) _start_http_server(ctx *AppContext, handler http.Handler, port int, name string) error {
	l := ctx.L()
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: handler,
	}
	if a.features.TLS.Enabled {
		tlsConfig, err := a.features.TLS.TLSConfig()
		if err != nil {
			l.Error("[Startup docs] encountered an error on startup", zap.Error(err))
			return err
		}
		srv.TLSConfig = tlsConfig
	}

	a.threadWg.Add(1)
	go func() {
		defer a.threadWg.Done()
		if a.features.TLS.Enabled {
			err := srv.ListenAndServeTLS("", "")
			if err != nil && err != http.ErrServerClosed {
				l.Error("[Startup http] encountered an error on startup",
					zap.Error(err), zap.String("name", name))
				return
			}
		} else {
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				l.Error("[Startup http] encountered an error on startup",
					zap.Error(err), zap.String("name", name))
				return
			}
		}
	}()

	a.stopServers = append(a.stopServers, func() (string, error) {
		return name, srv.Shutdown(ctx)
	})

	return nil
}

func (a *App) _startup_docs(ctx *AppContext) error {
	l := ctx.L()
	l.Info("[Startup docs] Serving htnl docs",
		zap.String("path", a.features.Docs.DocsPath), zap.String("api_path", a.features.Docs.DocsApiPath))
	if a.Features().HTTP.Enabled {
		docsFs := http.FS(os.DirFS(a.features.Docs.DocsPath))
		a.Features().HTTP.Mux.Handle(a.features.Docs.DocsApiPath, http.FileServer(docsFs))
	}

	if a.Features().Gin.Enabled {
		a.Features().Gin.Engine.StaticFS(a.features.Docs.DocsApiPath, http.FS(os.DirFS(a.features.Docs.DocsPath)))
	}

	a.state.DocsInitialized = true
	return nil
}

func (a *App) _startup_tls(ctx *AppContext) error {
	f := a.features.TLS
	l := ctx.L()
	cfg := &tls.Config{}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("failed to get system cert pool: %v", err)
	}

	if len(f.CABytes) > 0 {
		l.Info("[Startup TLS] using CA bytes in tls config")
		ok := pool.AppendCertsFromPEM(f.CABytes)
		if !ok {
			return fmt.Errorf("failed to append all ca bytes")
		}
	} else if f.CAFile != "" {
		l.Info("[Startup TLS] Using CA file for tls config", zap.String("ca", f.CAFile))
		b, err := os.ReadFile(f.CAFile)
		if err != nil {
			return fmt.Errorf("failed to read ca file %s: %v", f.CAFile, err)
		}

		ok := pool.AppendCertsFromPEM(b)
		if !ok {
			return fmt.Errorf("failed to add pem bytes from ca file %s", f.CAFile)
		}
	} else {
		l.Info("[Startup TLS] No CA given")
	}
	cfg.ClientCAs = pool
	cfg.RootCAs = pool

	var tlsCert tls.Certificate
	var keyBytes []byte
	var certBytes []byte

	if len(f.ServerCertBytes) > 0 {
		l.Info("[Startup TLS] server cert bytes given")
		certBytes = f.ServerCertBytes
	} else if f.ServerCertFile != "" {
		l.Info("[Startup TLS] server cert file given", zap.String("cert_file", f.ServerCertFile))
		certBytes, err = os.ReadFile(f.ServerCertFile)
		if err != nil {
			return fmt.Errorf("failed to read cert file %s: %v", f.ServerCertFile, err)
		}
	} else {
		return fmt.Errorf("no server cert given for TLS")
	}

	if len(f.ServerKeyBytes) > 0 {
		l.Info("[Startup TLS] server key bytes given")
		keyBytes = f.ServerKeyBytes
	} else if f.ServerKeyFile != "" {
		l.Info("[Startup TLS] server key file given", zap.String("key_file", f.ServerKeyFile))
		keyBytes, err = os.ReadFile(f.ServerKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file %s: %v", f.ServerKeyFile, err)
		}
	} else {
		return fmt.Errorf("no server key gven for tls")
	}

	if len(keyBytes) > 0 && len(certBytes) > 0 {
		l.Info("[Startup TLS] Loading key pair...")
		tlsCert, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return fmt.Errorf("failed to load key pair: %v", err)
		}
		cfg.Certificates = append(cfg.Certificates, tlsCert)
	}

	a.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: cfg,
		},
	}

	a.state.TLSInitialized = true
	return nil
}

func (a *App) _startup_logging_api(ctx *AppContext) error {
	l := ctx.L()

	l.Debug("[Startup Logging API] adding logging routes")
	handler := v1.Std()
	err := a._start_http_server(ctx, handler, a.features.LoggingAPI.Port, "logging-api")
	if err != nil {
		l.Error("[Startup Logging API] encountered an error on startup", zap.Error(err))
		return err
	}
	l.Debug("[Startup Logging API] finished adding logging routes")
	a.state.LoggingAPIInitialized = true
	return nil
}

func (a *App) _startup_redis(ctx *AppContext) error {
	l := ctx.L()

	l.Debug("[Startup Redis] initializing Redis client")
	err := redis.Init(a.features.Redis.redisDB)
	if err != nil {
		l.Error("[Startup Redis] failed to initialize Redis client", zap.Error(err))
		return err
	}

	a.state.RedisInitialized = true

	l.Debug("[Startup Redis] Redis client initialized successfully")
	return nil
}

func (a *App) _startup_valkey(ctx *AppContext) error {
	l := ctx.L()

	l.Debug("[Startup Valkey] initializing Valkey client")
	err := valkey.Init(&a.features.Valkey.valkeyDB)
	if err != nil {
		l.Error("[Startup Valkey] failed to initialize Valkey client", zap.Error(err))
		return err
	}

	a.state.ValkeyInitialized = true
	l.Debug("[Startup Valkey] Valkey client initialized successfully")
	return nil
}

func (a *App) _startup_rsa(ctx *AppContext) error {
	l := ctx.L()

	privKeyPath := a.features.RSA.PrivateKeyPath
	pubKeyPath := a.features.RSA.PublicKeyPath

	if privKeyPath != "" && pubKeyPath != "" {
		l.Debug("[Startup RSA] initializing RSA keys with paths",
			zap.String("rsa_private_key_path", privKeyPath),
			zap.String("rsa_pub_key_path", pubKeyPath),
		)
		if !fileExists(privKeyPath) {
			l.Info("[Startup RSA] RSA private key not found", zap.String("path", privKeyPath))
			return nil
		}

		if !fileExists(pubKeyPath) {
			l.Info("[Startup RSA] RSA public key not found", zap.String("path", pubKeyPath))
			return nil
		}
		privKey := mustReadFile(privKeyPath)
		pubKey := mustReadFile(pubKeyPath)

		if err := keys.InitRSA(privKey, pubKey); err != nil {
			return err
		}

		a.state.RSAInitialized = true
		l.Debug("[Startup RSA] RSA keys initialized successfully")
	} else {
		l.Debug("[Startup RSA] no RSA key paths provided")
		rsa, err := keys.NewRSA()
		if err != nil {
			return err
		}
		keys.SetRSA(rsa)
	}

	a.state.RSAInitialized = true
	return nil
}

func (a *App) _startup_jwt(ctx *AppContext) error {
	l := ctx.L()

	configs := a.features.JWT.tokenConfiguration
	configPaths := a.features.JWT.tokenConfigurationPaths
	privKeyPath := a.features.JWT.PrivateKeyPath
	pubKeyPath := a.features.JWT.PubKeyPath

	if privKeyPath != "" && pubKeyPath != "" {

		if !fileExists(privKeyPath) {
			l.Info("[Startup JWT] JWT private key does not exist", zap.String("path", privKeyPath))
			return nil
		}

		if !fileExists(pubKeyPath) {
			l.Info("[Startup JWT] JWT public key does not exist", zap.String("path", pubKeyPath))
			return nil
		}
		l.Debug("[Startup JWT] initializing JWT keys with paths",
			zap.String("jwt_private_key_path", privKeyPath),
			zap.String("jwt_pub_key_path", pubKeyPath),
		)
		jwtPrivKey := mustReadFile(privKeyPath)
		jwtPubKey := mustReadFile(pubKeyPath)
		if err := keys.InitJwt(jwtPrivKey, jwtPubKey); err != nil {
			return err
		}

		a.state.JWTInitialized = true
		l.Debug("[Startup JWT] JWT keys initialized successfully")
	} else {
		l.Debug("[Startup JWT] no JWT key paths provided, creating a key...")
		rsaKey, err := keys.NewRSA()
		if err != nil {
			return err
		}
		jwtKey := keys.NewJWTKey(*rsaKey)
		keys.SetJwt(jwtKey)
	}

	for _, configPath := range configPaths {
		if !fileExists(configPath) {
			l.Error("[Startup JWT] JWT token config file not found", zap.String("path", configPath))
			return fmt.Errorf("JWT token config file not found: %s", configPath)
		}

		config, err := jwt.ParseTokenConfigFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to parse token config file %s: %v", configPath, err)
		}
		ctx.issuerToTokenConfigs[config.Issuer] = *config
	}

	for _, cfg := range configs {
		ctx.issuerToTokenConfigs[cfg.Issuer] = cfg
	}

	a.state.JWTInitialized = true
	return nil
}

func (a *App) _startup_registry(ctx *AppContext) error {
	l := ctx.L()

	if a.features.Registry.registryPath != nil && *a.features.Registry.registryPath != "" {
		registryPath := *a.features.Registry.registryPath
		if !fileExists(registryPath) {
			l.Info("[Startup Registry] registry file not found", zap.String("registry_path", registryPath))
			return ErrRegistryFileNotFound
		}

		l.Debug("[Startup Registry] initializing registry with path", zap.String("path", registryPath))
		if err := registry.Init(registryPath); err != nil {
			l.Error("[Startup Registry] failed to initialize registry", zap.Error(err))
			return err
		}
		a.state.RegistryInitialized = true
		l.Debug("[Startup Registry] registry initialized successfully")

	} else {
		l.Debug("[Startup Registry] no registry path provided, using localhost")
		registry.InitLocalhost()
	}
	a.state.RegistryInitialized = true
	return nil

}

func (a *App) _startup_health(ctx *AppContext) error {
	l := ctx.L()
	l.Info("[Startup Health] initializing health with path", zap.String("path", a.features.Health.Path))
	port := 8080
	if a.features.Gin.Enabled {
		port = a.features.Gin.Port
		e := a.features.Gin.Engine
		e.GET(a.features.Health.Path, func(ctx *gin.Context) {
			if a.healthCheck != nil {
				if a.healthCheck() {
					ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
					a.state.Healthy = true
				} else {
					ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
					a.state.Healthy = false
				}
			} else {
				ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
				a.state.Healthy = true
			}
		})
	}

	if a.features.HTTP.Enabled {
		port = a.features.HTTP.Port
		a.features.HTTP.Mux.HandleFunc(a.features.Health.Path, func(w http.ResponseWriter, r *http.Request) {
			if a.healthCheck != nil {
				if a.healthCheck() {
					w.WriteHeader(http.StatusOK)
					a.state.Healthy = true
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					a.state.Healthy = false
				}
			} else {
				w.WriteHeader(http.StatusOK)
				a.state.Healthy = true
			}
		})
	}

	a.threadWg.Add(1)
	go func() {
		protocol := "http"
		if a.features.TLS.Enabled {
			protocol = "https"
		}

		defer a.threadWg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(a.features.Health.Interval) * time.Second):
				url := fmt.Sprintf("%s://localhost:%d%s",
					protocol,
					port, a.features.Health.Path)
				_, err := a.httpClient.Get(url)
				if err != nil {
					l.Error("[Startup Health] got an error from health check", zap.Error(err))
				}
			}
		}
	}()

	return nil

}

func (a *App) _run_gin(ctx *AppContext) error {
	l := a.l
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.features.Gin.Port),
		Handler: a.features.Gin.Engine,
	}

	if a.features.TLS.Enabled {
		tlsConfig, err := a.features.TLS.TLSConfig()
		if err != nil {
			l.Error("[Running Gin] Failed to get TLS config", zap.Error(err))
			return err
		}
		server.TLSConfig = tlsConfig
	}
	a.threadWg.Add(2)

	go func() {
		defer a.threadWg.Done()
		if a.features.TLS.Enabled {
			l.Debug("[Running Gin] Starting HTTPS server with Gin")
			err := server.ListenAndServeTLS("", "")
			if err != nil {
				l.Error("[Running Gin] encountered an error on listen and serve", zap.Error(err))
				return
			}
		} else {
			l.Debug("[Running Gin] Starting HTTP server with Gin")
			err := server.ListenAndServe()
			if err != nil {
				l.Error("[Running Gin] encountered an error on listen and serve", zap.Error(err))
				return
			}
		}
	}()
	go func() {
		defer a.threadWg.Done()
		<-ctx.Done()
		err := server.Shutdown(ctx)
		if err != nil {
			l.Error("[Running Gin] encountered an error on startup", zap.Error(err))
			return
		}
	}()

	a.state.GinInitialized = true
	return nil
}

func (a *App) _run_http(ctx *AppContext) error {
	l := a.l
	err := a._start_http_server(ctx, a.features.HTTP.Mux, a.features.HTTP.Port, "http")
	if err != nil {
		l.Error("[Running HTTP] encountered an error on startup", zap.Error(err))
		return err
	}

	a.state.HTTPInitialized = true
	return nil
}

func (a *App) _run(ctx *AppContext) error {
	l := a.l

	if a.features.Gin.Enabled {
		err := a._run_gin(ctx)
		if err != nil {
			a.l.Error("[Running Gin] encountered an error when running gin", zap.Error(err))
			return err
		}
	}

	if a.features.HTTP.Enabled {
		err := a._run_http(ctx)
		if err != nil {
			a.l.Error("[Running HTTP] encountered an error when running http", zap.Error(err))
			return err
		}
	}
	a.state.Running = true
	a.state.Healthy = true

	a.threadWg.Add(1)
	go func() {
		defer a.threadWg.Done()
		<-ctx.Done()
		for _, f := range a.stopServers {
			server, err := f()
			if err != nil {
				l.Error("[Startup] encountered an error when stopping server", zap.Error(err), zap.String("server", server))
			}

			l.Info("[Startup] stopped server", zap.String("server", server))
		}
	}()

	return nil
}

func (a *App) _startup(ctx context.Context) error {
	l := a.l
	if a.onPanic != nil {
		defer func() {
			if err := recover(); err != nil {
				l.Warn("recovered from panic", zap.Any("error", err))
				a.onPanic(err)
			}
		}()
	}

	startup_funcs := []func(ctx *AppContext) error{}

	if a.features.Registry.enabled {
		l.Info("[Startup] registry enabled")
		startup_funcs = append(startup_funcs, a._startup_registry)
	}

	if a.features.JWT.Enabled {
		l.Info("[Startup] JWT enabled")
		startup_funcs = append(startup_funcs, a._startup_jwt)
	}

	if a.features.RSA.Enabled {
		l.Info("[Startup] RSA enabled")
		startup_funcs = append(startup_funcs, a._startup_rsa)
	}

	if a.features.SQL.Enabled {
		l.Info("[Startup] SQL enabled")
		startup_funcs = append(startup_funcs, a._startup_sql)
	}

	if a.features.Redis.Enabled {
		l.Info("[Startup] Redis enabled")
		startup_funcs = append(startup_funcs, a._startup_redis)
	}

	if a.features.Docs.Enabled {
		l.Info("[Startup] Docs enabled")
		startup_funcs = append(startup_funcs, a._startup_docs)
	}

	if a.features.LoggingAPI.Enabled {
		l.Info("[Startup] Logging API enabled")
		startup_funcs = append(startup_funcs, a._startup_logging_api)
	}

	if a.features.TLS.Enabled {
		l.Info("[Startup] TLS enabled")
		startup_funcs = append(startup_funcs, a._startup_tls)
	}

	if a.features.Health.Enabled {
		l.Info("[Startup] Health enabled")
		startup_funcs = append(startup_funcs, a._startup_health)
	}

	if a.features.Valkey.Enabled {
		l.Info("[Startup] Valkey enabled")
		startup_funcs = append(startup_funcs, a._startup_valkey)
	}

	appCtx := NewAppContext(ctx, a.l)
	for _, f := range startup_funcs {
		err := f(appCtx)
		if err != nil {
			return err
		}
	}

	if a.setup != nil {
		l.Debug("[Startup] Running app...")
		err := a.setup(appCtx)
		if err != nil {
			l.Error("[Startup] encountered an error on setup", zap.Error(err))
			return err
		}
	}

	err := a._run(appCtx)
	if err != nil {
		l.Error("[Startup] encountered an error when running app", zap.Error(err))
		return err
	}

	if a.running != nil {
		a.threadWg.Add(1)
		go func() {
			defer a.threadWg.Done()
			l.Debug("[Startup] Running app...")
			err := a.running(appCtx)
			if err != nil {
				l.Error("[Startup] encountered an error during running", zap.Error(err))
			}
		}()
	}

	a.threadWg.Wait()
	l.Debug("[Startup] app stopped")
	a.state.Running = false
	if a.stopped != nil {
		err := a.stopped(appCtx)
		if err != nil {
			l.Error("[Startup] encountered an error on stopping", zap.Error(err))
			return err
		}
	}

	return nil
}
