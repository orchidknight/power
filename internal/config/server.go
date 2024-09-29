package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	ZeroBits        int
	Address         string
	ShutdownTimeout int
	HashCashTTL     int
}

func NewServerConfig() (*ServerConfig, error) {
	viper.AutomaticEnv()

	serverAddr := viper.GetString("SERVER_ADDRESS")
	if serverAddr == "" {
		return nil, fmt.Errorf("server address is empty")
	}

	zeroBits := viper.GetInt("SERVER_ZERO_BITS_COUNT")
	if zeroBits == 0 {
		return nil, fmt.Errorf("zero bits is empty")
	}

	shutdownTimeout := viper.GetInt("SERVER_SHUTDOWN_TIMEOUT")
	if shutdownTimeout == 0 {
		return nil, fmt.Errorf("shutdown timeout is empty")
	}

	hashCashTTL := viper.GetInt("SERVER_HASHCASH_TTL")
	if shutdownTimeout == 0 {
		return nil, fmt.Errorf("hashcash ttl is empty")
	}

	config := &ServerConfig{
		Address:         serverAddr,
		ZeroBits:        zeroBits,
		ShutdownTimeout: shutdownTimeout,
		HashCashTTL:     hashCashTTL,
	}

	return config, nil
}

func (cfg ServerConfig) String() string {
	return fmt.Sprintf("serverAddress: %s zeroBits: %d shutdownTimeout: %d hashcashTTL: %d", cfg.Address, cfg.ZeroBits, cfg.ShutdownTimeout, cfg.HashCashTTL)
}
