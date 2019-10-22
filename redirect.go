package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type (
	Redirect struct {
		Host     string
		Replacer *strings.Replacer
		Logger   logrus.FieldLogger

		Options RedirectOptions
	}

	RedirectOptions struct {
		StatusCodeRedirect int
		Schema             string
		KeepURI            bool
	}
)

func NewRedirect(l logrus.FieldLogger, v *viper.Viper) Redirect {
	r := strings.NewReplacer(
		v.GetString("redirect.keyword.slash"), "/",
		v.GetString("redirect.keyword.dot"), ".",
		v.GetString("redirect.keyword.colon"), ":",
		v.GetString("redirect.keyword.interrogation-mark"), "?",
		v.GetString("redirect.keyword.ampersand"), "&",
		v.GetString("redirect.keyword.equal"), "=",
		v.GetString("redirect.keyword.percent"), "%",
	)

	sc := http.StatusTemporaryRedirect
	if v.GetBool("redirect.options.permanent_redirect") {
		sc = http.StatusPermanentRedirect
	}

	s := "https://"
	if !v.GetBool("redirect.options.enforce_https") {
		s = "" // keep as requested
	}

	return Redirect{
		Host:     v.GetString("redirect.host"),
		Replacer: r,
		Logger:   l,
		Options: RedirectOptions{
			StatusCodeRedirect: sc,
			Schema:             s,
			KeepURI:            v.GetBool("redirect.options.keep_uri"),
		},
	}
}

func (r Redirect) Redirect(req *http.Request) (string, int, error) {
	// remove port, if any (can't resolve with)
	h := strings.TrimRight(req.Host, ":0123456789")
	cname, err := net.LookupCNAME(h)
	if err != nil {
		return "", 0, err
	}

	r.Logger.WithFields(logrus.Fields{"host": h, "cname": cname}).Debug("resolve")

	sd := strings.TrimSuffix(cname, r.Host)
	t := fmt.Sprintf("%s%s", r.Options.Schema, r.Replacer.Replace(sd))

	if r.Options.KeepURI && req.RequestURI != "/" {
		t = fmt.Sprintf("%s%s", t, req.RequestURI)
	}

	return t, r.Options.StatusCodeRedirect, nil
}
