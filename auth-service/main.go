package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// App struct (para injeção de dependência)
type App struct {
	DB        *sql.DB
	MasterKey string
}

func main() {
	// Carrega o .env para desenvolvimento local. Em produção, isso não fará nada.
	_ = godotenv.Load()

	// --- Configuração ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001" // Porta padrão
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL deve ser definida")
	}

	masterKey := os.Getenv("MASTER_KEY")
	if masterKey == "" {
		log.Fatal("MASTER_KEY deve ser definida")
	}

	// --- Conexão com o Banco ---
	db, err := connectDB(databaseURL)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	app := &App{
		DB:        db,
		MasterKey: masterKey,
	}

	// --- Rotas da API ---
	mux := http.NewServeMux()

	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "togglemaster_http_requests_total",
			Help: "Total de requisicoes HTTP do Auth Service.",
		},
		[]string{"route", "method", "code"},
	)

	durations := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "togglemaster_http_request_duration_seconds",
			Help:    "Duracao das requisicoes HTTP em segundos.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "method"},
	)

	prometheus.MustRegister(requests, durations)

	instrument := func(route string, handler http.Handler) http.Handler {
		labels := prometheus.Labels{"route": route}

		return promhttp.InstrumentHandlerDuration(
			durations.MustCurryWith(labels),
			promhttp.InstrumentHandlerCounter(
				requests.MustCurryWith(labels),
				handler,
			),
		)
	}

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", app.healthHandler)

	// Endpoint público para validar uma chave
	mux.Handle(
		"/validate",
		instrument("/validate", http.HandlerFunc(app.validateKeyHandler)),
	)

	// Endpoints de "admin" para criar/gerenciar chaves
	// Eles são protegidos pelo middleware de autenticação
	mux.Handle(
		"/admin/keys",
		instrument(
			"/admin/keys",
			app.masterKeyAuthMiddleware(
				http.HandlerFunc(app.createKeyHandler),
			),
		),
	)

	log.Printf("Serviço de Autenticação (Go) rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// connectDB inicializa e testa a conexão com o PostgreSQL
func connectDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Conectado ao PostgreSQL com sucesso!")
	return db, nil
}
