package moneybot

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config store configuration params
type Config struct {
	DbFile                     string
	LogIntoFile                bool
	LogSQL                     bool
	TelegramToken              string
	GRPCServer                 string
	NotificationTimeout        time.Duration
	MonobankIntegrationEnabled bool
	MonobankWebhookUrl         string
	MonobankToken              string
	MonobankChatId             int32
}

// InitConfig init configurations from file and .env
func InitConfig() (*Config, error) {
	v := viper.New()

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	v.SetDefault("db_file", "db.sqlite3")
	v.SetDefault("enable_file_log", true)
	v.SetDefault("enable_sql_log", true)
	v.SetDefault("notification_timeout", 10)
	v.SetDefault("monobank_integration", false)

	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := &Config{}
	config.NotificationTimeout = time.Duration(v.GetInt("notification_timeout")) * time.Second
	config.LogIntoFile = v.GetBool("enable_file_log")
	config.LogSQL = v.GetBool("enable_sql_log")
	config.DbFile = v.GetString("db_file")
	config.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	config.GRPCServer = os.Getenv("GRPC_SERVER_ADDRESS")

	config.MonobankIntegrationEnabled = v.GetBool("monobank_integration")
	if config.MonobankIntegrationEnabled {
		config.MonobankWebhookUrl = os.Getenv("MONOBANK_WEBHOOK_URL")
		config.MonobankToken = os.Getenv("MONOBANK_TOKEN")
		chatId, err := strconv.Atoi(os.Getenv("MONOBANK_CHAT_ID"))
		if err != nil {
			return nil, err
		}
		config.MonobankChatId = int32(chatId)

	}

	return config, nil
}
