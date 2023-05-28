package util

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if _, err := os.Stat(path + "/.env"); err == nil {
		viper.AddConfigPath(path)
		viper.SetConfigName(".env")
		viper.SetConfigType("env")

		err = viper.ReadInConfig()
		if err != nil {
			return config, err
		}
	}

	config = Config{
		EmailSenderName:      viper.GetString("EMAIL_SENDER_NAME"),
		EmailSenderAddress:   viper.GetString("EMAIL_SENDER_ADDRESS"),
		EmailSenderPassword:  viper.GetString("EMAIL_SENDER_PASSWORD"),
	}

	return config, nil
}