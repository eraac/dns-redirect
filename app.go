package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

type (
	App struct {
		Server http.Server
		Logger logrus.FieldLogger

		Redirect Redirect
	}
)

func NewApp(l logrus.FieldLogger, v *viper.Viper) *App {
	return &App{
		Server: http.Server{
			Addr: fmt.Sprintf(":%d", v.GetInt("app.http.port")),
		},
		Logger:   l,
		Redirect: NewRedirect(l, v),
	}
}

func (a *App) Listen() error {
	a.Logger.WithField("port", a.Server.Addr).Info("http server listening")
	defer a.Logger.WithField("port", a.Server.Addr).Info("http server stopping")

	return a.Server.ListenAndServe()
}

func (a *App) RegisterHandler() {
	h := http.NewServeMux()

	h.HandleFunc("/health_check", a.health)
	h.HandleFunc("/", a.redirect)

	a.Server.Handler = h
}

func (a *App) Close(ctx context.Context) error {
	a.Logger.Info("graceful shutdown...")

	return a.Server.Shutdown(ctx)
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	t, sc, err := a.Redirect.Redirect(r)
	if err != nil {
		a.Logger.WithField("domain", r.Host).Error(err)
		http.Error(w, "resolver fail", http.StatusInternalServerError)
		return
	}

	a.Logger.WithFields(logrus.Fields{"domain": r.Host, "target": t}).Info("redirect")

	http.Redirect(w, r, t, sc)
}

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	a.Logger.Debug("health")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}
