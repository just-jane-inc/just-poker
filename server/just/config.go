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

	Pepper []byte
	pepper string `env:"JUST_POKER_PEPPER,required"`
}

var (
	Logger     = jlog.NewLogger("just-poker").WithConsole(jlog.DEBUG)
	Env        Config
	DBConnPool *pgxpool.Pool
)

func init() {
	Logger.Info("initializing config...")
	_ = godotenv.Load()
	var err error
	err = env.Parse(&Env)
	if err != nil {
		log.Fatalf("unable to parse ennvironment variables: %e", err)
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
