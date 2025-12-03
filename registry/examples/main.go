package main

import (
	"os"

	"github.com/ooqls/getset/registry"
	"gopkg.in/yaml.v3"
)

func main() {
	r := registry.Registry{
		Kafka: &registry.MessageBroker{
			Server: registry.Server{
				Name: "kafka",
				Host: "https://localhost",
				Port: 10,
				TLS: &registry.TLSConfig{
					Enabled:  true,
					CertPath: "cert.pem",
					KeyPath:  "key.pem",
					CaPath:   "ca.pem",
				},
				Auth: registry.Auth{
					Enabled:  true,
					Username: "user",
					Password: "password",
				},
			},
			Topics: []string{"abc"},
		},
		Redis: &registry.Database{
			Server: registry.Server{
				Name: "redis",
				Host: "https://redis.com",
				Port: 9001,
			},
			Database: "redis",
		},
		Postgres: &registry.Database{
			Server: registry.Server{
				Name: "pg",
				Host: "https://localhost",
				Port: 9999,
				TLS: &registry.TLSConfig{
					Enabled: false,
				},
				Auth: registry.Auth{
					Enabled:  true,
					Username: "admin",
					Password: "postgres",
				},
			},
			Database: "auth",
		},
	}

	b, err := yaml.Marshal(r)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("example.yaml", b, 0666); err != nil {
		panic(err)
	}

}
