package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Defines the structure for allowed file types.
type FileTypes struct {
	Plain     string   `mapstructure:"plain"`
	Images    []string `mapstructure:"images"`
	Documents []string `mapstructure:"documents"`
}

// Configures file storage settings.
type File struct {
	UploadDir  string    `mapstructure:"upload_dir"`
	StorageDir string    `mapstructure:"storage_dir"`
	Soffice    string    `mapstructure:"soffice"`
	Types      FileTypes `mapstructure:"types"`
}

// Configures MongoDB connection parameters.
type Mongo struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DbName   string `mapstructure:"dbName"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// Constructs the MongoDB connection string.
func (m *Mongo) MongoURI() string {
	if m.User != "" && m.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
			m.User, m.Password, m.Host, m.Port, m.DbName)
	}
	return fmt.Sprintf("mongodb://%s:%s/%s", m.Host, m.Port, m.DbName)
}

type Email struct {
	Host *string `mapstructure:"host"`
	User *string `mapstructure:"user"`
	Pass *string `mapstructure:"pass"`
}

// Oauth configures OAuth secrets.
type Oauth struct {
	YandexSecret string `mapstructure:"yandex_secret"`
	VkID         string `mapstructure:"vk_id"`
	VkSecret     string `mapstructure:"vk_secret"`
}

// Telegram configures Telegram bot settings (uses pointers for conditional enabling).
type Telegram struct {
	ChatID *int    `mapstructure:"chat_id"`
	Token  *string `mapstructure:"token"`
}

// Agenda configures the scheduler.
type Agenda struct {
	DbCollection string `mapstructure:"dbCollection"`
	Pooltime     string `mapstructure:"pooltime"`
	Concurrency  int    `mapstructure:"concurrency"`
}

// API configures API settings.
type API struct {
	Prefix string `mapstructure:"prefix"`
}

// Config holds the application's configuration.
type Config struct {
	Env        string `mapstructure:"env"`
	Port       int    `mapstructure:"port"`
	LogLevel   string `mapstructure:"log_level"`
	Secret     string `mapstructure:"secret"`      // For JWT
	SaltRounds int    `mapstructure:"salt_rounds"` // For bcrypt

	EnableEmail    bool `mapstructure:"enable_email"`    // Control flag
	EnableTelegram bool `mapstructure:"enable_telegram"` // Control flag

	Mongo    Mongo    `mapstructure:"mongo"`
	Email    Email    `mapstructure:"email"`
	Oauth    Oauth    `mapstructure:"oauth"`
	Telegram Telegram `mapstructure:"telegram"`
	Agenda   Agenda   `mapstructure:"agenda"`
	API      API      `mapstructure:"api"`
	File     File     `mapstructure:"file"`
}

// global config variable
var Cfg *Config

// Load configuration from environment variables and potentially a .env file.
func Load(paths ...string) (*Config, error) {
	v := viper.New()

	// --- Set Defaults ---
	v.SetDefault("env", "development")
	v.SetDefault("api.prefix", "/api")
	v.SetDefault("log_level", "info")

	// --- Configure Viper for Environment Variables ---
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// --- Configure Viper for .env File ---
	if len(paths) > 0 {
		for _, path := range paths {
			v.AddConfigPath(path)
		}
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}
	v.SetConfigName(".env")
	v.SetConfigType("env")

	// Attempt to read the .env file
	err := v.ReadInConfig()
	// Handle specific error if .env MUST exist (like in your TS code)
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return nil, fmt.Errorf("⚠️ couldn't find .env file: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// --- Unmarshal into Struct ---
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if !cfg.EnableEmail {
		cfg.Email.Host = nil
		cfg.Email.User = nil
		cfg.Email.Pass = nil
	}
	if !cfg.EnableTelegram {
		cfg.Telegram.ChatID = nil
		cfg.Telegram.Token = nil
	}

	// --- Assign to global or return ---
	Cfg = &cfg
	return &cfg, nil
}
