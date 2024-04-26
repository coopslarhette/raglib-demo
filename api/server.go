package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"raglib/internal/retrieval"
	"syscall"
)

type Server struct {
	router             *chi.Mux
	qdrantPointsClient qdrant.PointsClient
	serpAPIClient      *retrieval.SERPAPIClient
	openAIClient       *openai.Client
}

func NewServer(conn *grpc.ClientConn, openAIClient *openai.Client) *Server {
	s := &Server{
		router:             chi.NewRouter(),
		qdrantPointsClient: qdrant.NewPointsClient(conn),
		serpAPIClient:      retrieval.NewSerpApiClient(os.Getenv("SERPAPI_API_KEY"), http.DefaultClient),
		openAIClient:       openAIClient,
	}

	s.useMiddleWare()
	s.establishRoutes()

	return s
}

func (s *Server) Start(ctx context.Context) {
	port := 5000
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.router,
	}

	shutdownComplete := handleShutdown(func() {
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server.Shutdown failed: %v\n", err)
		}
	})

	if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		<-shutdownComplete
	} else {
		log.Printf("http.ListenAndServe failed: %v\n", err)
	}

	log.Println("Shutdown gracefully")
}

func handleShutdown(onShutdownSignal func()) <-chan struct{} {
	shutdown := make(chan struct{})

	go func() {
		shutdownSignal := make(chan os.Signal, 1)
		signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

		<-shutdownSignal

		onShutdownSignal()
		close(shutdown)
	}()

	return shutdown
}

func (s *Server) useMiddleWare() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
}

func (s *Server) establishRoutes() {
	s.router.Get("/health", healthHandler)
	s.router.Get("/search", s.searchHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	type HealthResponse struct {
		Status string `json:"status"`
	}

	response := HealthResponse{
		Status: "OK",
	}

	render.JSON(w, r, response)
}
