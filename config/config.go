package config

import (
	"github.com/eshop/pkg/envutils"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// Configs contain all the configs data in cluding default.
type Configs struct {
	Env                 string `envconfig:"ENV" default:"development"`
	HTTPPort            string `envconfig:"HTTPPort" default:":8080"`
	DBConString         string `envconfig:"MongoDB Con String" default:"mongodb://localhost:27017/"`
	DBName              string `envconfig:"MongoDB Name" default:"bigbasket"`
	RedisConString      string `envconfig:"Redis Con String" default:"localhost:6379"`
	JWTAccessSecretKey  string `envconfig:"JWT Access Secret Key" default:"jdnfksdmfksd"`
	JWTRefreshSecretKey string `envconfig:"JWT Refresh Secret Key" default:"mcmvmkmsdnfsdmfdsjf"`
}

// GetConfigs get the configs details.
func GetConfigs() (*Configs, error) {
	cfg := &Configs{}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "get configs")
	}

	err = checkConfigEnv(cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "check configs")
	}
	return cfg, nil
}

func checkConfigEnv(config *Configs) error {
	err := envutils.Check(config.Env)
	if err != nil {
		return errors.WithMessage(err, "environment")
	}
	return nil
}
