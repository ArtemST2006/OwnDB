package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ArtemST2006/OwnDB/backend/internal/http/handler"
	"github.com/ArtemST2006/OwnDB/backend/internal/repository"
	"github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	repository := repository.NewRepository()
	handler := handler.NewHandler(repository)

	srv := new(Server)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := srv.Run("8000", handler.InitRoutes()); err != nil {
			logrus.Fatalf("main.go/main/error in init http server: %s", err.Error())
		}
	}()

	wg.Wait()

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Fatalf("main.go/main/error with shutting down %s", err.Error())
	}
}
