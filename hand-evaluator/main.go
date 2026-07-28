package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

var (
	Env Config
)

type Card struct {
	Rank rune `json:"rank"`
	Suit rune `json:"suit"`
}

type Response struct {
	Error      string `json:"error"`
	Evaluation int    `json:"evaluation"`
}

func main() {
	log.Print("starting...")

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /evaluate", EvaluateHand)

	rootMux := http.NewServeMux()
	rootMux.Handle(
		"/",
		IgnoreTrailingSlash(apiMux),
	)

	server := &http.Server{
		Addr:    ":" + Env.Port,
		Handler: rootMux,
	}

	log.Printf("starting server on :%s", Env.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("encountered error in server: %v", err)
	}
}

func EvaluateHand(w http.ResponseWriter, r *http.Request) {
	log.Print("request received to evaluate hand")

	var hand []Card
	if err := json.NewDecoder(r.Body).Decode(&hand); err != nil {
		WriteJSONResponse(w, 400, Response{Error: err.Error()})
		return
	}

	cards := make([]string, len(hand))
	for i, card := range hand {
		cards[i] = fmt.Sprintf("%c%c", card.Rank, card.Suit)
	}

	args := strings.Join(cards, " ")
	log.Printf("getting hand score for: %s", args)
	cmd := exec.Command(Env.PokerEvalCLI, cards...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Run()
	outString := stdout.String()
	errString := stderr.String()

	outString = strings.Trim(outString, "\n")
	log.Printf("evaluating hand: %s [out: %s] [err: %s]", args, outString, errString)

	if errString != "" {
		WriteJSONResponse(w, 400, Response{Error: errString})
		return
	}

	outInt, err := strconv.Atoi(outString)
	if err != nil {
		WriteJSONResponse(w, 400, Response{Error: err.Error()})
		return
	}

	WriteJSONResponse(w, 200, Response{Evaluation: outInt})
}

func WriteJSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encountered error when writing json response: %s", err.Error())
	}
}

func IgnoreTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimRight(r.URL.Path, "/")
			r.URL.RawPath = strings.TrimRight(r.URL.RawPath, "/")
		}

		next.ServeHTTP(w, r)
	})
}

type Config struct {
	ElasticHost     string `env:"ELASTIC_HOST,required"`
	ElasticKey      string `env:"ELASTIC_API_KEY,required"`
	ElasticLogIndex string `env:"ELASTIC_LOG_INDEX,required"`
	DiscordLogHook  string `env:"DISCORD_ERR_LOGS_HOOK,required"`
	Port            string `env:"JUST_POKER_PORT,required"`
	PokerEvalCLI    string `env:"POKER_EVAL_CLI,required"`
}

func init() {
	log.Print("initializing config...")
	_ = godotenv.Load()
	if err := env.Parse(&Env); err != nil {
		log.Printf("unable to parse ennvironment variables: %e", err)
		panic(err)
	}
}
