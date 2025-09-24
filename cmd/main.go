package main

import (
	"log"
	"net/http"
	"time"

	_config "bitbucket.org/edts/go-task-management/config"
	_db "bitbucket.org/edts/go-task-management/internal/db"
	"bitbucket.org/edts/go-task-management/internal/graph/_generated"
	_directives "bitbucket.org/edts/go-task-management/internal/graph/directives"
	_dl "bitbucket.org/edts/go-task-management/internal/graph/loaders"
	_resolver "bitbucket.org/edts/go-task-management/internal/graph/resolver"
	_mw "bitbucket.org/edts/go-task-management/internal/middleware"
	_pubsub "bitbucket.org/edts/go-task-management/internal/pubsub"
	_repo "bitbucket.org/edts/go-task-management/internal/repository"
	_usecase "bitbucket.org/edts/go-task-management/internal/usecase"
	_customErr "bitbucket.org/edts/go-task-management/pkg/errors"
	_logger "bitbucket.org/edts/go-task-management/pkg/logger"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

var logs = _logger.GetContextLoggerf(nil)

func init() {
	// Load config
	_config.LoadConfig()
	// Use directive custom message validator
	defaultTranslation()
}

func defaultTranslation() {
	// This one if you want to add custom message based on field
	//_directives.ValidateAddTranslation("email", " not a valid email (custom message)")
}

func main() {
	// Init db connection
	dbConn, err := _db.NewDatabase(&_config.AppConfigInstance.Database)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer dbConn.Close()

	// Retrieve port from config
	port := _config.AppConfigInstance.App.Port
	if port == "" {
		port = defaultPort
	}

	repo := _repo.NewRepository(dbConn)
	pubsub := _pubsub.NewPubSub()
	uc := _usecase.NewUsecase(repo, pubsub)
	dataloader := _dl.NewLoaders(repo)
	resolver := _resolver.NewResolver(uc, dataloader)

	// Generated config
	genConf := _generated.Config{Resolvers: resolver}
	// Use directives binding for validator
	genConf.Directives.Binding = _directives.Binding

	//TODO: add AuthDirective

	// Init GraphQL server
	srv := handler.New(_generated.NewExecutableSchema(genConf))
	// Use custom error presenter
	srv.SetErrorPresenter(_customErr.CustomErrorPresenter)
	// Use custom recover function
	srv.SetRecoverFunc(_customErr.CustomRecoverFunc)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Subscriptions handler
	subscriptionSrv := handler.New(_generated.NewExecutableSchema(genConf))
	subscriptionSrv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	})
	subscriptionSrv.AddTransport(transport.SSE{})

	// Health check route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_mw.CORSHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, r)
	})

	http.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	http.Handle("/graphql", _mw.RequestMiddleware(_mw.CORSHandler(_dl.Middleware(repo, srv))))
	http.Handle("/subscription", _mw.RequestMiddleware(_mw.CORSHandler(_dl.Middleware(repo, subscriptionSrv))))

	logs.Infof("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
