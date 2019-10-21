package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stderr)
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	file := flag.String("config", "config.yaml", "config file")
	flag.Parse()

	v, err := LoadConfiguration(*file)
	if err != nil {
		logrus.WithField("context", "load_configuration").Fatal(err)
	}

	l := LoadLogger(v)

	app := NewApp(l, v)
	app.RegisterHandler()

	go func() {
		if err := app.Listen(); err != nil {
			l.WithField("context", "listen").Fatal(err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Close(ctx); err != nil {
		l.WithField("context", "close").Fatal(err)
	}
}

func LoadConfiguration(filename string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(filename)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	logrus.WithField("filename", viper.ConfigFileUsed()).Info("reading config file")

	return v, nil
}

func LoadLogger(v *viper.Viper) logrus.FieldLogger {
	l := logrus.New()
	l.SetLevel(logrus.Level(v.GetInt("app.log_level")))
	l.SetFormatter(&logrus.TextFormatter{})
	l.SetOutput(os.Stderr)

	return l
}
