package registry

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var r *Registry
var m sync.Mutex = sync.Mutex{}

func Set(reg Registry) {
	m.Lock()
	defer m.Unlock()

	r = &reg
}

func Get() *Registry {
	m.Lock()
	defer m.Unlock()
	if r == nil {
		panic("please init registry before using it")
	}

	return r
}

func InitLocalhost() {
	m.Lock()
	defer m.Unlock()

	rlocalhost := Registry{
		Kafka: &MessageBroker{
			Server: Server{
				Host: "localhost",
				Port: 9092,
			},
		},
		Nats: &MessageBroker{
			Server: Server{
				Host: "localhost",
				Port: 4222,
			},
		},
		Redis: &Database{
			Server: Server{
				Host: "localhost",
				Port: 6379,
			},
			Database: "0",
		},
		Postgres: &Database{
			Server: Server{
				Host: "localhost",
				Port: 5432,
				Auth: Auth{
					Enabled:  true,
					Username: "postgres",
					Password: "postgres",
				},
			},
			Database: "postgres",
		},
		Elasticsearch: &Database{
			Server: Server{
				Host: "localhost",
				Port: 9200,
				Auth: Auth{
					Enabled:  true,
					Username: "elastic",
					Password: "elastic",
				},
			},
		},
		Mongo: &Database{
			Server: Server{
				Host: "localhost",
				Port: 27017,
			},
			Database: "mongo",
		},
	}
	r = &rlocalhost
}

func InitDefault() error {
	m.Lock()
	defer m.Unlock()

	p := os.Getenv("REGISTRY_PATH")
	if p == "" {
		p = "/opt/config/registry.yaml"
	}

	return Init(p)
}

func Init(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to init registry: %v", err)
	}

	var reg Registry
	err = yaml.Unmarshal(b, &reg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config into registry: %v", err)
	}

	r = &reg

	return nil
}
