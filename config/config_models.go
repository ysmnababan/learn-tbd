package config

type Config struct {
	App AppConfig

	DB DBConfig
}

type AppConfig struct {
	Name     string
	Version  string
	Port     int
	Schema   string
	URL      string
	LogLevel string
}

type DBConfig struct {
	Host     string
	Username string
	Password string
	Port     string
	Name     string

	SSLMode  string
	TimeZone string

	MaxIdleConns int
	MaxOpenConns int
	LogLevel     string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string // AKA access key
	Password string // AKA secret key
	Sender   string
}
