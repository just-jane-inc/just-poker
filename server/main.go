package main

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/just-jane-inc/just-poker/server/game"
	"github.com/just-jane-inc/just-poker/server/just"
	"github.com/just-jane-inc/just-poker/server/user"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
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

		_, _ = w.Write(openAPISpec)

	default:
		http.NotFound(w, r)
	}
}

// @title           BAHMS Poker Tournament
// @version         1.0
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
// @host      localhost:8080
// @BasePath  /api
func main() {
	just.Logger.Debug("starting server...")

	apiMux := http.NewServeMux()

	apiMux.HandleFunc(
		"GET /user/{userID}",
		user.OnGetUserRequest,
	)

	apiMux.HandleFunc(
		"POST /game",
		game.OnCreateGame,
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
		"POST /game/{game_id}/rack/exchange",
		game.OnExchangeChips,
	)

	// create a new user
	apiMux.HandleFunc(
		"POST /user",
		user.OnCreateUser,
	)

	// deletes the user who provides the request token
	apiMux.HandleFunc(
		"DELETE /user/me",
		user.OnDeleteMe,
	)

	// get all bots for a twitch user
	apiMux.HandleFunc(
		"GET /user/twitch/{twitch_id}",
		user.OnGetUsers,
	)

	apiMux.HandleFunc(
		"DELETE /user/{user_id}/key",
		user.OnRenewKey,
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
    
    just.Logger.Info("just__started")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
