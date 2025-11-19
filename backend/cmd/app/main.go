package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func (s *Server) Shutdown(ctx context.Context, repo *repository.Repository) error {
	if err := repo.Flush(); err != nil {
		return err
	}

	return s.httpServer.Shutdown(ctx)
}

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if ok := repository.Settings(); !ok {
		logrus.Fatal("main.go/main/error in init repo")
	}

	var user_index, data_index map[string]int64
	user_index, data_index = repository.ParseToIndex()

	repository := repository.NewRepository(user_index, data_index)
	handler := handler.NewHandler(repository)

	srv := new(Server)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Run("8000", handler.InitRoutes()); err != nil {
			logrus.Fatalf("main.go/main/error in init http server: %s", err.Error())
		}
	}()

	<-quit

	if err := srv.Shutdown(context.Background(), repository); err != nil {
		logrus.Fatalf("main.go/main/error with shutting down %s", err.Error())
	}
}
