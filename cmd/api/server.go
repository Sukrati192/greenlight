package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(app.logger, "", 0),
	}
	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.PrintInfo("caught signal", map[string]string{"signal": s.String()})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			shutdownErr <- err
		}
		app.logger.PrintInfo("completing background tasks", map[string]string{"addr": srv.Addr})
		app.wg.Wait()
		shutdownErr <- nil
	}()
	app.logger.PrintInfo("starting server ", map[string]string{"addr": srv.Addr, "env": app.config.env})
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownErr
	if err != nil {
		return err
	}
	app.logger.PrintInfo("stopped server", map[string]string{"addr": srv.Addr})

	return nil
}
