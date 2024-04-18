package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var config *Config
var rateLimitConfig map[string]*RateLimitConfig

type Config struct {
	Server     Server
	Cache      CacheConfig
	RateLimits []RateLimitConfig
}

type Server struct {
	Port string
}

type RateLimitConfig struct {
	AppId     string
	IpAddress string
	Threshold int
	ApiPath   string
	Interval  RateLimitIntervalConfig
}

type RateLimitIntervalConfig struct {
	Value    int
	Type     RateLimitIntervalType
	TimeUnit TimeUnit
}

type CacheConfig struct {
	Host string
	Port string
}

type RateLimitIntervalType string

type TimeUnit string

const (
	Fixed    RateLimitIntervalType = "FIXED"
	Variable RateLimitIntervalType = "VARIABLE"
)

const (
	Second TimeUnit = "SECOND"
	Minute TimeUnit = "MINUTE"
	Hour   TimeUnit = "HOUR"
	DAY    TimeUnit = "DAY"
)

func InializeConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config/" + os.Getenv("ENV") + "/" + "config.yml")
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config, %s", err)
	}
	viper.Unmarshal(&config)
	setRateLimitConfig()
}

func GetConfig() *Config {
	return config
}

func GetRateLimitConfig() map[string]*RateLimitConfig {
	return rateLimitConfig
}

func setRateLimitConfig() {
	rateLimitConfig = map[string]*RateLimitConfig{}
	for i, rateLimit := range config.RateLimits {
		if rateLimit.AppId != "" && validRateLimitConfig(rateLimit) {
			rateLimitConfig[rateLimit.AppId] = &config.RateLimits[i]
		}
		if rateLimit.IpAddress != "" && validRateLimitConfig(rateLimit) {
			rateLimitConfig[rateLimit.IpAddress] = &config.RateLimits[i]
		}
	}
}

func validRateLimitConfig(rateLimit RateLimitConfig) bool {
	return rateLimit.Threshold > 0 && rateLimit.ApiPath != "" && rateLimit.Interval.Value > 0
}
