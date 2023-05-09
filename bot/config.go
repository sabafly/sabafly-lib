package botlib

import (
	"encoding/gob"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"

	"github.com/sabafly/sabafly-lib/db"

	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v2"
)

func LoadConfig(config_filepath string) (*Config, error) {
	file, err := os.Open(config_filepath)
	if os.IsNotExist(err) {
		if file, err = os.Create(config_filepath); err != nil {
			return nil, err
		}
		switch filepath.Ext(config_filepath) {
		case ".json":
			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "\t")
			err = encoder.Encode(defaultConfig)
		case ".yml", ".yaml":
			err = yaml.NewEncoder(file).Encode(file)
		case ".toml":
			err = toml.NewEncoder(file).SetArraysMultiline(true).SetIndentSymbol("\t").SetIndentTables(true).Encode(defaultConfig)
		case ".xml":
			encoder := xml.NewEncoder(file)
			encoder.Indent("", "\t")
			err = encoder.Encode(defaultConfig)
		case ".gob":
			err = gob.NewEncoder(file).Encode(defaultConfig)
		default:
			panic("unknown config file type " + filepath.Ext(config_filepath))
		}
		if err != nil {
			return nil, err
		}
		return nil, errors.New("config file not found, created new one")
	} else if err != nil {
		return nil, err
	}

	var cfg Config
	switch filepath.Ext(config_filepath) {
	case ".json":
		err = json.NewDecoder(file).Decode(&cfg)
	case ".yml", ".yaml":
		err = yaml.NewDecoder(file).Decode(&cfg)
	case ".tml", ".toml":
		err = toml.NewDecoder(file).Decode(&cfg)
	case ".xml":
		err = xml.NewDecoder(file).Decode(&cfg)
	case ".gob":
		err = gob.NewDecoder(file).Decode(&cfg)
	default:
		panic("unknown config file type" + filepath.Ext(config_filepath))
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

var defaultConfig = Config{
	DevMode:            false,
	DevOnly:            false,
	DevGuildIDs:        make([]snowflake.ID, 0),
	DevUserIDs:         make([]snowflake.ID, 0),
	LogLevel:           "INFO",
	Token:              "YOUR TOKEN HERE",
	DMPermission:       false,
	ShouldSyncCommands: true,
	DBConfig: db.DBConfig{
		Host: "localhost",
		Port: "6379",
		DB:   0,
	},
	Dislog: DislogConfig{
		Enabled:        false,
		WebhookChannel: 0,
		WebhookID:      0,
		WebhookToken:   "YOUR WEBHOOK TOKEN HERE",
	},
}

type Config struct {
	DevMode            bool           `json:"dev_mode"`
	DevOnly            bool           `json:"dev_only"`
	DevGuildIDs        []snowflake.ID `json:"dev_guild_id"`
	DevUserIDs         []snowflake.ID `json:"dev_user_id"`
	LogLevel           string         `json:"log_level"`
	Token              string         `json:"token"`
	DMPermission       bool           `json:"dm_permission"`
	ShouldSyncCommands bool           `json:"sync_commands"`
	DBConfig           db.DBConfig    `json:"db_config"`
	Dislog             DislogConfig   `json:"dislog"`
	ClientID           snowflake.ID   `json:"client_id"`
	Secret             string         `json:"secret"`
}

type DislogConfig struct {
	Enabled        bool         `json:"enabled"`
	WebhookChannel snowflake.ID `json:"webhook_channel"`
	WebhookID      snowflake.ID `json:"webhook_id"`
	WebhookToken   string       `json:"webhook_token"`
}
