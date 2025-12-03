package jwt

import (
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func ParseTokenConfigFile(f string) (*TokenConfiguration, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var cfg TokenConfiguration
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

type TokenConfiguration struct {
	Audience                []string `yaml:"audience"`
	Issuer                  string   `yaml:"issuer"`
	IdGenType               string   `yaml:"id_gen_type"`
	ValidityDurationSeconds float64  `yaml:"validity_duration_seconds"`
}

func (tc *TokenConfiguration) GenerateId() string {
	defaultIdGen := uuid.NewString
	var id string
	switch tc.IdGenType {
	case "uuid":
		id = uuid.NewString()
	default:
		id = defaultIdGen()
	}

	return id
}
