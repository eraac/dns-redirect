package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type (
	Redirect struct {
		host     string
		logger   logrus.FieldLogger
		resolver Resolver

		options RedirectOptions
	}

	RedirectOptions struct {
		replacer           *strings.Replacer
		statusCodeRedirect int
		schema             string
		keepURI            bool
	}
)

func NewRedirect(l logrus.FieldLogger, v *viper.Viper) (*Redirect, error) {
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

	resolver, err := NewResolverFromConfig(l, v)
	if err != nil {
		l.WithField("context", "new_resolver_from_config").Error(err)
		return nil, err
	}

	return &Redirect{
		host:     v.GetString("redirect.host"),
		logger:   l,
		resolver: resolver,
		options: RedirectOptions{
			replacer:           r,
			statusCodeRedirect: sc,
			schema:             s,
			keepURI:            v.GetBool("redirect.options.keep_uri"),
		},
	}, nil
}

func (r Redirect) Redirect(req *http.Request) (string, int, error) {
	// remove port, if any (can't resolve with)
	o := strings.TrimRight(req.Host, ":0123456789")
	cname, err := r.resolver.Resolve(o)
	if err != nil {
		return "", 0, err
	}

	r.logger.WithFields(logrus.Fields{"host": r.host, "origin": o, "cname": cname}).Debug("resolver response")

	sd := strings.TrimSuffix(cname, r.host)
	t := fmt.Sprintf("%s%s", r.options.schema, r.options.replacer.Replace(sd))

	if r.options.keepURI && req.RequestURI != "/" {
		t = fmt.Sprintf("%s%s", t, req.RequestURI)
	}

	return t, r.options.statusCodeRedirect, nil
}
