package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/jakka/minimule-backend/graph/resolver"
	"github.com/jakka/minimule-backend/internal/auth"
	"github.com/jakka/minimule-backend/internal/cache"
	"github.com/jakka/minimule-backend/internal/config"
	"github.com/jakka/minimule-backend/internal/database"
	"github.com/jakka/minimule-backend/internal/database/queries"
	"github.com/jakka/minimule-backend/internal/middleware"
	"github.com/jakka/minimule-backend/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Infrastructure ─────────────────────────────────────────────────────────
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		slog.Error("database connect", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := cache.Connect(ctx, cfg)
	if err != nil {
		slog.Error("redis connect", "err", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// ── Services ───────────────────────────────────────────────────────────────
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	q := queries.New(db)

	authSvc := service.NewAuthService(q, jwtSvc)
	cartSvc := service.NewCartService(q)
	root := &resolver.RootResolver{
		Auth:         authSvc,
		ProductSvc:   service.NewProductService(q),
		CategorySvc:  service.NewCategoryService(q),
		CartSvc:      cartSvc,
		OrderSvc:     service.NewOrderService(q, cartSvc),
		PostSvc:      service.NewPostService(q, authSvc),
		ReviewSvc:    service.NewReviewService(q),
		PaymentSvc:   service.NewPaymentService(q),
		ShippingSvc:  service.NewShippingService(q),
		PromotionSvc: service.NewPromotionService(q),
		NotifSvc:     service.NewNotificationService(q),
		SearchSvc:    service.NewSearchService(q),
	}

	// ── GraphQL schema ─────────────────────────────────────────────────────────
	schemaBytes, err := os.ReadFile("graph/schema/schema.graphqls")
	if err != nil {
		slog.Error("read schema", "err", err)
		os.Exit(1)
	}
	schema := graphql.MustParseSchema(string(schemaBytes), root,
		graphql.MaxDepth(cfg.QueryDepthLimit),
		graphql.MaxParallelism(10),
		graphql.UseStringDescriptions(),
	)

	// ── HTTP server ────────────────────────────────────────────────────────────
	mux := http.NewServeMux()
	mux.Handle("/graphql", &relay.Handler{Schema: schema})
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	if cfg.GraphQLPlayground {
		mux.HandleFunc("/playground", playgroundHandler)
		slog.Info("GraphQL playground enabled", "url", "http://localhost:"+cfg.Port+"/playground")
	}

	// Chain middleware: CORS → Auth → RateLimit → mux
	handler := middleware.CORS(cfg.CORSOrigins)(
		middleware.Auth(jwtSvc)(
			middleware.RateLimit(redisClient, cfg)(mux),
		),
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ── Graceful shutdown ──────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	slog.Info("server stopped")
}

// playgroundHandler serves a minimal GraphiQL playground.
func playgroundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>miniMule GraphQL Playground</title>
  <link rel="stylesheet" href="https://unpkg.com/graphiql/graphiql.min.css"/>
</head>
<body style="margin:0">
  <div id="graphiql" style="height:100vh"></div>
  <script crossorigin src="https://unpkg.com/react/umd/react.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/react-dom/umd/react-dom.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/graphiql/graphiql.min.js"></script>
  <script>
    const fetcher = GraphiQL.createFetcher({ url: '/graphql' });
    ReactDOM.render(React.createElement(GraphiQL, { fetcher }), document.getElementById('graphiql'));
  </script>
</body>
</html>`))
}
