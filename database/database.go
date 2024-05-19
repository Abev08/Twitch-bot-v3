package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

type secrets uint16

const (
	// Twitch
	TwitchName secrets = iota
	TwitchCustomerID
	TwitchPassword
	// TikTok
	TikTokSessionID
	// Spotify
	SpotifyClientID
	SpotifyClientSecret
	// Discord
	DiscrodClientID
	DiscordClientSecret
	DiscordBotToken
	DiscordChannelID
)

func (c secrets) toString() string {
	switch c {
	case TwitchName:
		return "TwitchName"
	case TwitchCustomerID:
		return "TwitchCustomerID"
	case TwitchPassword:
		return "TwitchPassword"
	case TikTokSessionID:
		return "TikTokSessionID"
	case SpotifyClientID:
		return "SpotifyClientID"
	case SpotifyClientSecret:
		return "SpotifyClientSecret"
	case DiscrodClientID:
		return "DiscrodClientID"
	case DiscordClientSecret:
		return "DiscordClientSecret"
	case DiscordBotToken:
		return "DiscordBotToken"
	case DiscordChannelID:
		return "DiscordChannelID"
	default:
		return ""
	}
}

type config uint16

const (
	ChannelName config = iota
	// Notification types
	Follow
	Subscription
	SubscriptionGift
	SubscriptionGiftReceived
	Cheer
	Raid
	Timeout
	Ban
)

func (c config) toString() string {
	switch c {
	case ChannelName:
		return "ChannelName"
	case Follow:
		return "Follow"
	case Subscription:
		return "Subscription"
	case SubscriptionGift:
		return "SubscriptionGift"
	case SubscriptionGiftReceived:
		return "SubscriptionGiftReceived"
	case Cheer:
		return "Cheer"
	case Raid:
		return "Raid"
	case Timeout:
		return "Timeout"
	case Ban:
		return "Ban"
	default:
		return ""
	}
}

var secretsData = make(map[secrets]string, DiscordChannelID+1)
var configData = make(map[config]string)

var initialized bool
var dbConn *sql.DB

func Init() {
	if initialized {
		return
	}
	initialized = true

	slog.Info("Initializing database")
	var err error

	// Connection
	dbConn, err = sql.Open("sqlite3", ".db")
	if err != nil {
		panic(err) // This has to work, if it didn't the driver is broken?
	}
	// defer dbConn.Close() This closes connection at the end of Init method - too early

	// Try to create required tables, it will return an error if the table already exists
	dbConn.Exec("CREATE TABLE secrets (name TEXT NOT NULL UNIQUE, value TEXT);")
	dbConn.Exec("CREATE TABLE config (name TEXT NOT NULL UNIQUE, value TEXT);")

	// Get or create secrets data
	if secretsData[TwitchName], err = getValueOrCreteNew("secrets", "TwitchName", ""); err != nil {
		panic(err)
	}
	if secretsData[TwitchCustomerID], err = getValueOrCreteNew("secrets", "TwitchCustomerID", ""); err != nil {
		panic(err)
	}
	if secretsData[TwitchPassword], err = getValueOrCreteNew("secrets", "TwitchPassword", ""); err != nil {
		panic(err)
	}
	if secretsData[TikTokSessionID], err = getValueOrCreteNew("secrets", "TikTokSessionID", ""); err != nil {
		panic(err)
	}
	if secretsData[SpotifyClientID], err = getValueOrCreteNew("secrets", "SpotifyClientID", ""); err != nil {
		panic(err)
	}
	if secretsData[SpotifyClientSecret], err = getValueOrCreteNew("secrets", "SpotifyClientSecret", ""); err != nil {
		panic(err)
	}
	if secretsData[DiscrodClientID], err = getValueOrCreteNew("secrets", "DiscrodClientID", ""); err != nil {
		panic(err)
	}
	if secretsData[DiscordClientSecret], err = getValueOrCreteNew("secrets", "DiscordClientSecret", ""); err != nil {
		panic(err)
	}
	if secretsData[DiscordBotToken], err = getValueOrCreteNew("secrets", "DiscordBotToken", ""); err != nil {
		panic(err)
	}
	if secretsData[DiscordChannelID], err = getValueOrCreteNew("secrets", "DiscordChannelID", ""); err != nil {
		panic(err)
	}

	// Get or create config data
	if configData[ChannelName], err = getValueOrCreteNew("config", "ChannelName", ""); err != nil {
		panic(err)
	}
}

func Close() {
	dbConn.Close()
}

func IsRequiredInfoProvided() bool {
	// Minimum required info is Twitch Bot stuff and channel name
	var ok = true
	ok = ok && len(secretsData[TwitchName]) != 0
	ok = ok && len(secretsData[TwitchCustomerID]) != 0
	ok = ok && len(secretsData[TwitchPassword]) != 0

	ok = ok && len(configData[ChannelName]) != 0
	return ok
}

func getValueOrCreteNew(table, valueName, defaultValue string) (string, error) {
	if dbConn == nil {
		return "", errors.New("database connection not establised")
	}

	var err error
	var val string
	err = dbConn.QueryRow(fmt.Sprintf("SELECT value FROM %s WHERE name='%s' LIMIT 1;", table, valueName)).Scan(&val)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// Create new record
			_, err = dbConn.Exec(fmt.Sprintf("INSERT INTO %s (name, value) VALUES ('%s', '%s');", table, valueName, defaultValue))
			if err != nil {
				return "", err
			}
			val = defaultValue
		} else {
			return "", err
		}
	}

	return val, nil
}

func GetSecretsValue(key secrets) string {
	return secretsData[key]
}

func GetConfigValue(key config) string {
	return configData[key]
}

// Updates secrets key value in the database.
func UpdateSecretsValue(key secrets, value string) error {
	secretsData[key] = value
	return updateValueInDatabase("secrets", key.toString(), value)
}

// Updates config key value in the database.
func UpdateConfigValue(key config, value string) error {
	configData[key] = value
	return updateValueInDatabase("config", key.toString(), value)
}

func updateValueInDatabase(table, key, newValue string) error {
	slog.Info("Database row update", "table", table, "key", key)
	_, err := dbConn.Exec(fmt.Sprintf("UPDATE %s SET Value='%s' WHERE name='%s';", table, newValue, key))
	return err
}

func GetSecretsAndConfigAsJson() []byte {
	var data = make(map[string]interface{})
	data["Secrets"] = []string{
		secretsData[TwitchName],
		secretsData[TwitchCustomerID],
		secretsData[TwitchPassword],
		secretsData[TikTokSessionID],
		secretsData[SpotifyClientID],
		secretsData[SpotifyClientSecret],
		secretsData[DiscrodClientID],
		secretsData[DiscordClientSecret],
		secretsData[DiscordBotToken],
		secretsData[DiscordChannelID],
	}

	data["Config"] = []string{
		configData[ChannelName],
	}

	if d, e := json.Marshal(data); e == nil {
		return d
	}
	return nil
}
