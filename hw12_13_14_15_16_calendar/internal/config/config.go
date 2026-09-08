package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf
	HTTP    HTTPConf
	Storage StorageConf
}

type LoggerConf struct {
	Level string
}

type HTTPConf struct {
	Host string
	Port string
}

type StorageConf struct {
	Type string // memory или sql
	DSN  string // строка подключения к БД (для sql)
}

func NewConfig(filePath string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return Config{}, fmt.Errorf("ошибка чтения конфига: %w", err)
	}

	return config, nil
}
