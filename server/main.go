package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/just-jane-inc/just-poker/server/game"
	"github.com/just-jane-inc/just-poker/server/just"
	"github.com/just-jane-inc/just-poker/server/user"
)

//go:embed docs/swagger.json
var openAPISpec []byte

const scalarHTML = `<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<meta
			name="viewport"
			content="width=device-width, initial-scale=1"
		>
		<title>Game API</title>
	</head>

	<body>
		<div id="app"></div>

		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>

		<script>
			const basePath = window.location.pathname.endsWith("/")
				? window.location.pathname
				: window.location.pathname + "/"

			Scalar.createApiReference("#app", {
				url: basePath + "openapi.json",
				layout: "classic"
			})
		</script>
	</body>
</html>`

var (
	buildSpecOnce sync.Once
	servedSpec    []byte
)

func servedOpenAPISpec() []byte {
	buildSpecOnce.Do(func() {
		servedSpec = openAPISpec

		var spec map[string]any
		if err := json.Unmarshal(openAPISpec, &spec); err != nil {
			just.Logger.Warn("unable to parse openapi spec, serving unmodified")
			return
		}

		spec["servers"] = []any{
			map[string]any{
				"url": just.Env.SwaggerServerURL,
			},
		}

		out, err := json.Marshal(spec)
		if err != nil {
			just.Logger.Warn("unable to rewrite openapi servers, serving spec unmodified")
			return
		}

		servedSpec = out
	})

	return servedSpec
}

func documentationHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.URL.Path {
	case "/swagger", "/swagger/":
		w.Header().Set(
			"Content-Type",
			"text/html; charset=utf-8",
		)

		_, _ = w.Write([]byte(scalarHTML))

	case "/swagger/openapi.json":
		w.Header().Set(
			"Content-Type",
			"application/json; charset=utf-8",
		)

		_, _ = w.Write(servedOpenAPISpec())

	default:
		http.NotFound(w, r)
	}
}

// @title           BAHMS Poker Tournament
// @version         1.0
// @servers.url   https://game.bahms.org/api/poker
// @servers.description	production api idk whatever you want to have the description here......yeaaaaaaah, exactly umm and now the game
// @description.markdown
// @termsOfService  http://swagger.io
// @securityDefinitions.bearerauth BearerAuth
// @name Authorization
// @in header
// @description Type "Bearer" followed by a space and your token (e.g. "Bearer your_token_here").
// @contact.name   Red_Epicness
// @contact.url    http://swagger.io
// @license.name  BAHMS
// @license.url   http://license.bahms.org
// @BasePath  /api
func main() {
	just.Logger.Info("starting server...")

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/game/{game_id}/state/ws", game.OnCreateGameConnection)
	apiMux.HandleFunc("POST /game/{game_id}/state/listen", game.OnRegisterListener)
	apiMux.HandleFunc("GET /game/{game_id}/state/listen", game.OnGetNextListenerEvent)

	apiMux.HandleFunc(
		"GET /game",
		game.OnGetCurrentActiveGames,
	)

	apiMux.HandleFunc(
		"POST /game",
		game.OnCreateGame,
	)

	apiMux.HandleFunc(
		"DELETE /game/{game_id}",
		game.OnDeleteGame,
	)

	apiMux.HandleFunc(
		"POST /game/{game_id}/hand",
		game.OnStartNextHand,
	)

	apiMux.HandleFunc(
		"POST /game/{game_id}/action",
		game.OnPlayerAction,
	)

	apiMux.HandleFunc(
		"POST /game/{game_id}/player",
		game.OnJoinGameRequest,
	)

	apiMux.HandleFunc(
		"GET /game/{game_id}/state",
		game.OnGetCurrentGameState,
	)

	apiMux.HandleFunc(
		"POST /game/{game_id}/started",
		game.OnStartGame,
	)

	apiMux.HandleFunc(
		"POST /game/{game_id}/state",
		game.OnStartGameFromState,
	)

	apiMux.HandleFunc(
		"POST /game/{game_id}/chip/exchange",
		game.OnExchangeChips,
	)

	// deletes the user who provides the request token
	apiMux.HandleFunc(
		"DELETE /user/me",
		user.OnDeleteMe,
	)

	// =====Admin Panel=====
	// These have no auth, internal only.

	// create a new user
	apiMux.HandleFunc(
		"POST /user",
		user.OnCreateUser,
	)

	// get all bots for a twitch user
	apiMux.HandleFunc(
		"GET /user/twitch/{twitch_id}",
		user.OnGetUsers,
	)

	// renew a key for a user
	apiMux.HandleFunc(
		"DELETE /user/{user_id}/key",
		user.OnRenewKey,
	)

	// delete a user
	apiMux.HandleFunc(
		"DELETE /user/{user_id}",
		user.OnDeleteUser,
	)

	// =====Admin Panel End=====

	apiMux.HandleFunc(
		"POST /hand-evaluator/evaluate",
		game.OnEvalHand,
	)

	rootMux := http.NewServeMux()

	rootMux.HandleFunc(
		"/swagger",
		documentationHandler,
	)
	rootMux.HandleFunc(
		"/swagger/",
		documentationHandler,
	)

	rootMux.Handle(
		"/",
		just.IgnoreTrailingSlash(apiMux),
	)

	server := &http.Server{
		Addr:    ":" + just.Env.Port,
		Handler: rootMux,
	}

	fmt.Println("just__started")

	fmt.Printf("listening on :%s\n", just.Env.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
