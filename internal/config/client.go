package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type ClientConfig struct {
	ServerAddress string
	MaxTries      int
}

func NewClientConfig() (*ClientConfig, error) {
	viper.AutomaticEnv()

	serverAddr := viper.GetString("CLIENT_SERVER_ADDRESS")
	if serverAddr == "" {
		return nil, fmt.Errorf("server address in empty")
	}

	maxTries := viper.GetInt("CLIENT_MAX_TRIES")
	if maxTries == 0 {
		return nil, fmt.Errorf("max tries is zero")
	}

	config := &ClientConfig{
		ServerAddress: serverAddr,
		MaxTries:      maxTries,
	}

	return config, nil
}

func (cfg ClientConfig) String() string {
	return fmt.Sprintf("serverAddress: %s maxTries: %d", cfg.ServerAddress, cfg.MaxTries)
}
