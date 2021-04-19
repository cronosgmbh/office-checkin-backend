package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

// Config contains all settings related to the actual service
//goland:noinspection ALL
type Config struct {
	Service struct {
		Port        string `yaml:"port", envconfig:"SERVER_PORT"`
		Environment string `yaml:"environment", envconfig:"SERVER_ENVIRONMENT"`
		LogLevel    string `yaml:"log_level", envconfig:"SERVER_LOG_LEVEL"`
		TaskInterval string `yaml:"task_interval", envconfig:"SERVER_TASK_INTERVAL"`
	} `yaml:"service"`
	MongoDB struct {
		Host       string `yaml:"host", envconfig:"MONGO_HOST"`
		Username   string `yaml:"username", envconfig:"MONGO_USERNAME"`
		Password   string `yaml:"password", envconfig:"MONGO_PASSWORD"`
		Database   string `yaml:"database", envconfig:"MONGO_DATABASE"`
		Collection string `yaml:"collection", envconfig:"MONGO_COLLECTION"`
	} `yaml:"mongodb"`
	Email struct {
		Host     string `yaml:"host", envconfig:"EMAIL_HOST"`
		Port     string `yaml:"port", envconfig:"EMAIL_PORT"`
		Username string `yaml:"username", envconfig:"EMAIL_USERNAME"`
		Password string `yaml:"password", envconfig:"EMAIL_PASSWORD"`
		FromName string `yaml:"from_name", envconfig:"EMAIL_FROM_NAME"`
		FromMail string `yaml:"from_mail"", envconfig:"EMAIL_FROM_MAIL"`
	} `yaml:"email"`
	Badge struct {
		BackgroundColor string `yaml:"background_color", envconfig:"BADGE_BG_COLOR"`
		ForegroundColor string `yaml:"foreground_color", envconfig:"BADGE_FORE_COLOR"`
	} `yaml:"badge"`
	Bookings struct {
		AutoDelete      bool   `yaml:"auto_delete", envconfig:"BOOKINGS_AUTO_DELETE"`
		DeleteAfterDays int    `yaml:"delete_after_days", envconfig:"BOOKINGS_DELETE_AFTER_DAYS"`
	} `yaml:"bookings"`
	Visitors struct {
		AutoDelete	    bool   `yaml:"auto_delete", envconfig:"VISITORS_AUTO_DELETE"`
		DeleteAfterDays	int    `yaml:"delete_after_days", envconfig:"VISITORS_DELETE_AFTER_DAYS"`
	} `yaml:"visitors"`
}

// loadConfig loads the config from config.yaml and env variables
func loadConfig(cfg *Config) {
	parseFile(cfg)
}

// parseFile reads the config.yaml config file
func parseFile(cfg *Config) {
	f, err := os.Open("config.yaml")
	if err != nil {
		logrus.Error(err)
		logrus.Info("please create a config.yaml file in the root of your project and make it readable for the service")
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		logrus.Error(err)
		os.Exit(5)
	}

}

// parseEnv reads all config data from environment variables and writes it into the struct
func parseEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		logrus.Warn(err)
	}
}
