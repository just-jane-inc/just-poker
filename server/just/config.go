// Package just the core package handling things that are critical to business logic but not directly associated to poker
package just

import (
	"context"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/elastic/go-elasticsearch/v8"
	jlog "github.com/just-jane-inc/just-services/just-logger/client"
)

type Config struct {
	Port         string `env:"JUST_POKER_PORT,required"`
	DBConnString string `env:"PG_URL,required"`
	PokerEvalURL string `env:"HAND_EVAL_ROUTE,required"`

	ElasticHost      string `env:"ELASTIC_HOST"`
	ElasticAPIKey    string `env:"ELASTIC_API_KEY"`
	ElasticIndexRoot string `env:"ELASTIC_INDEX_ROOT"`

	ElasticLoggingIndex string `env:"ELASTIC_LOG_INDEX"`
	FileLoggingFile     string `env:"FILE_LOG_PATH"`
	ConsoleLogLevel     string `env:"CONSOLE_LOG_LEVEL"`
	ElasticLogLevel     string `env:"ELASATIC_LOG_LEVEL"`
	FileLogLevel        string `env:"FILE_LOG_LEVEL"`

	Pepper []byte
	pepper string `env:"JUST_POKER_PEPPER,required"`
}

var (
	Logger     = jlog.NewLogger("just-poker")
	Env        Config
	DBConnPool *pgxpool.Pool
)

func getLogLevel(lvl string) jlog.JustLevel {
	switch lvl {
	case "TRACE":
		return jlog.TRACE
	case "DEBUG":
		return jlog.DEBUG
	case "INFO":
		return jlog.INFO
	case "WARN":
		return jlog.WARN
	case "ERROR":
		return jlog.ERROR
	case "FATAL":
		return jlog.FATAL
	case "PANIC":
		return jlog.PANIC
	default:
		return jlog.ERROR
	}
}

func init() {
	_ = godotenv.Load()
	var err error
	err = env.Parse(&Env)
	if err != nil {
		log.Fatalf("unable to parse ennvironment variables: %e", err)
	}

	Logger.WithConsole(getLogLevel(Env.ConsoleLogLevel))
	Logger.Info("initializing config...")

	if Env.ElasticLoggingIndex != "" {
		Logger.Info("enabling elastic logger...")
		Logger = Logger.WithElastic(Env.ElasticHost, Env.ElasticAPIKey, Env.ElasticLoggingIndex, getLogLevel(Env.ElasticLogLevel))
	}

	if Env.FileLoggingFile != "" {
		Logger.Info("enabling file logger...")
		Logger = Logger.WithFile(Env.FileLoggingFile, getLogLevel(Env.FileLogLevel))
	}

	Env.Pepper = []byte(Env.pepper)
	DBConnPool, err = pgxpool.New(context.Background(), Env.DBConnString)
	if err != nil {
		panic(err)
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			Env.ElasticHost,
		},
		APIKey: Env.ElasticAPIKey,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	RecordingHub = EventRecordingHub{
		elasticClient: client,
	}
}
