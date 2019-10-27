package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type (
	App struct {
		server http.Server
		logger logrus.FieldLogger

		redirect *Redirect
	}
)

func NewApp(l logrus.FieldLogger, v *viper.Viper) (*App, error) {
	r, err := NewRedirect(l, v)
	if err != nil {
		l.WithField("context", "new_redirect").Error(err)
		return nil, err
	}

	return &App{
		server: http.Server{
			Addr: fmt.Sprintf(":%d", v.GetInt("app.http.port")),
		},
		logger:   l,
		redirect: r,
	}, nil
}

func (a *App) Listen() error {
	a.logger.WithField("port", a.server.Addr).Info("http server listening")
	defer a.logger.WithField("port", a.server.Addr).Info("http server stopping")

	return a.server.ListenAndServe()
}

func (a *App) RegisterHandler() {
	h := http.NewServeMux()

	h.HandleFunc("/health_check", a.health)
	h.HandleFunc("/", a.redirection)

	a.server.Handler = h
}

func (a *App) Close(ctx context.Context) error {
	a.logger.Info("graceful shutdown...")

	return a.server.Shutdown(ctx)
}

func (a *App) redirection(w http.ResponseWriter, r *http.Request) {
	l, sc, err := a.redirect.Redirect(r)
	if err != nil {
		a.logger.WithField("domain", r.Host).Error(err)
		http.Error(w, "resolver fail", http.StatusInternalServerError)
		return
	}

	a.logger.WithFields(logrus.Fields{"domain": r.Host, "location": l}).Info("redirection")

	http.Redirect(w, r, l, sc)
}

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	a.logger.Debug("health")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}
