package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	LogLevel              string
	ZoneID                string
	AuthorizationKey      string
	NetworkInterface      string
	DomainNames           []string
	IPv4DDNS              bool
	IPv6DDNS              bool
	UpdateIntervalSeconds int
}

const defaultUpdateIntervalSeconds = 120

func Load() (*Config, error) {
	err := godotenv.Overload()
	if err != nil {
		logrus.Warnf(".env file not found: %v", err)
	}

	cfg := &Config{
		LogLevel:              strings.ToUpper(clean(os.Getenv("LOG_LEVEL"))),
		ZoneID:                clean(os.Getenv("ZONE_ID")),
		AuthorizationKey:      clean(os.Getenv("AUTHORIZATION_KEY")),
		NetworkInterface:      clean(os.Getenv("NETWORK_INTERFACE")),
		DomainNames:           parseDomainNames(os.Getenv("DOMAIN_NAMES")),
		IPv4DDNS:              os.Getenv("IPV4_DDNS") == "1",
		IPv6DDNS:              os.Getenv("IPV6_DDNS") == "1",
		UpdateIntervalSeconds: parseInterval(os.Getenv("UPDATE_INTERVAL_SECONDS")),
	}

	return cfg, nil
}

func Validate(cfg *Config) error {
	if cfg.LogLevel != "" && !isValidLogLevel(cfg.LogLevel) {
		return validationError("配置校验不通过: log_level不是有效的日志级别名，可选值：DEBUG, INFO, WARN, ERROR, FATAL, PANIC")
	}
	if cfg.ZoneID == "" {
		return validationError("配置校验不通过: zone_id不能为空")
	}
	if cfg.AuthorizationKey == "" {
		return validationError("配置校验不通过: authorization_key不能为空")
	}
	if cfg.NetworkInterface == "" {
		return validationError("配置校验不通过: network_interface不能为空")
	}
	if len(cfg.DomainNames) == 0 {
		return validationError("配置校验不通过: domain_names需要配置一个以上的域名")
	}
	return nil
}

func validationError(message string) error {
	logrus.Error(message)
	return errors.New(message)
}

func isValidLogLevel(level string) bool {
	switch level {
	case "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC":
		return true
	default:
		return false
	}
}

func clean(value string) string {
	return strings.TrimSpace(value)
}

func parseDomainNames(value string) []string {
	var names []string
	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func parseInterval(value string) int {
	if value == "" {
		return defaultUpdateIntervalSeconds
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return defaultUpdateIntervalSeconds
	}
	return seconds
}
