package config

import (
	"os"
	"reflect"
)

type Config struct {
	MongodbURL string `required:"true" env:"MONGODB_URL"`

	Port string `required:"true" env:"PORT"`
}

func GetEnvOrDefault(key string, defaultValue string) string {
	return getEnv(key, defaultValue)
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func LoadConfig(cfg interface{}) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		envVar := fieldType.Tag.Get("env")
		requiredVar := fieldType.Tag.Get("required")

		value := getEnv(envVar, "")
		field.SetString(value)

		if requiredVar == "true" && value == "" {
			panic("Missing required environment variable: " + envVar)
		}
	}
}

func NewConfig() *Config {
	cfg := &Config{}
	LoadConfig(cfg)
	return cfg
}
