package just

import (
	"context"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	jlog "github.com/just-jane-inc/just-services/just-logger/client"
)

type Config struct {
	ElasticHost     string `env:"ELASTIC_HOST,required"`
	ElasticKey      string `env:"ELASTIC_API_KEY,required"`
	ElasticLogIndex string `env:"ELASTIC_LOG_INDEX,required"`
	DiscordLogHook  string `env:"DISCORD_ERR_LOGS_HOOK,required"`
	Port            string `env:"JUST_POKER_PORT,required"`
	DbConnString    string `env:"PG_URL,required"`
	PokerEvalURL    string `env:"HAND_EVAL_ROUTE,required"`
	Pepper          []byte

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

	DBConnPool, err = pgxpool.New(context.Background(), Env.DbConnString)
	if err != nil {
		panic(err)
	}
}
