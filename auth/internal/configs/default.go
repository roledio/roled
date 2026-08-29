package configs

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/karrick/tparse/v2"
	"github.com/roledio/roled/pkg/constants"
	"github.com/spf13/viper"
)

type DefaultConfig struct {
	BaseURL        string `mapstructure:"base_url"`
	ConsoleBaseURL string `mapstructure:"console_base_url"`
	Env            string `mapstructure:"env"`
	Port           int64  `mapstructure:"port"`
	WebUseMinified bool   `mapstructure:"web_use_minified"`
	GTMContainerID string `mapstructure:"gtm_container_id"`
	DB             struct {
		Name     string `mapstructure:"name"`
		Host     string `mapstructure:"host"`
		Port     int64  `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"db"`
	Email struct {
		From string `mapstructure:"from"`
		SMTP struct {
			Host     string `mapstructure:"host"`
			Port     int64  `mapstructure:"port"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		} `mapstructure:"smtp"`
	} `mapstructure:"email"`
	Newrelic struct {
		LicenseKey string `mapstructure:"license_key"`
		Enabled    bool   `mapstructure:"enabled"`
	} `mapstructure:"newrelic"`
	CORS struct {
		AllowedDomains []string `mapstructure:"allowed_domains"`
	} `mapstructure:"cors"`
	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Prefix   string `mapstructure:"prefix"`
	} `mapstructure:"redis"`
	EncryptionMasterKey string `mapstructure:"encryption_master_key"`
	JWT                 struct {
		SigningKey                 string `mapstructure:"signing_key"`
		ExpiryDuration             string `mapstructure:"expiry_duration"`
		RefreshTokenExpiryDuration string `mapstructure:"refresh_token_expiry_duration"`
		AuthCodeExpiryDuration     string `mapstructure:"auth_code_expiry_duration"`
	} `mapstructure:"jwt"`
	CacheDefaultTTL                string `mapstructure:"cache_default_ttl"`
	CacheDefaultTTLDuration        time.Duration
	VerifyEmailExpiryDuration      string `mapstructure:"verify_email_expiry_duration"`
	ResetWithContextExpiryDuration string `mapstructure:"reset_password_expiry_duration"`
	ActivateMemberExpiryDuration   string `mapstructure:"activate_member_expiry_duration"`
	Upload                         struct {
		MaxFileSizeMB int    `mapstructure:"max_file_size_mb"`
		Driver        string `mapstructure:"driver"`
		Local         struct {
			UploadPath string `mapstructure:"upload_path"`
		} `mapstructure:"local"`
		S3 struct {
			AccessKey string `mapstructure:"access_key"`
			SecretKey string `mapstructure:"secret_key"`
			Region    string `mapstructure:"region"`
			Bucket    string `mapstructure:"bucket"`
			BaseURL   string `mapstructure:"base_url"`
			Endpoint  string `mapstructure:"endpoint"`
		} `mapstructure:"s3"`
	} `mapstructure:"upload"`
}

func (d *DefaultConfig) IsEnvProd() bool {
	return d.Env == constants.EnvProduction
}

func (d *DefaultConfig) IsEnvLocal() bool {
	return d.Env == constants.EnvLocal
}

func LoadDefaultConfig(path string) (*DefaultConfig, error) {
	if path == "" {
		return nil, errors.New("config path is empty")
	}

	// Load() the .env file to override the yaml config (for development convenience).
	// It will not override the config if the env variable is already set in the system.
	// Ignore the error since it's fine if the .env file does not exist (e.g. in production environment).
	// We will rely on the system env variables in that case.
	_ = godotenv.Load()

	v := viper.New()

	// Find and read the default config file
	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	// Read the env variables if exist to override the config file values
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read the config files with the settings above
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Unmarshal the config file into the struct
	var c DefaultConfig
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	// Expose base URLs as environment variables for migrations
	if os.Getenv("BASE_URL") == "" {
		err := os.Setenv("BASE_URL", c.BaseURL)
		fmt.Println("Error setting base url:", err)
	}
	if os.Getenv("CONSOLE_BASE_URL") == "" {
		err := os.Setenv("CONSOLE_BASE_URL", c.ConsoleBaseURL)
		fmt.Println("Error setting console base url:", err)
	}

	now := time.Now()
	expiryDuration, err := tparse.AddDuration(now, c.JWT.ExpiryDuration)
	if err != nil {
		return nil, err
	}
	refreshTokenExpiryDuration, err := tparse.AddDuration(now, c.JWT.RefreshTokenExpiryDuration)
	if err != nil {
		return nil, err
	}
	// Validate refresh token expiry duration must be greater than JWT expiry duration
	if expiryDuration.After(refreshTokenExpiryDuration) {
		return nil, fmt.Errorf("jwt.refresh_token_expiry_duration (%s) must be greater than jwt.expiry_duration (%s)",
			c.JWT.RefreshTokenExpiryDuration, c.JWT.ExpiryDuration)
	}
	c.CacheDefaultTTLDuration, err = tparse.AbsoluteDuration(now, c.CacheDefaultTTL)
	if err != nil {
		// Default to 5 minutes if parsing fails
		c.CacheDefaultTTLDuration = 5 * time.Minute
	}
	return &c, nil
}
