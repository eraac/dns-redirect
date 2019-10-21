package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type (
	Redirect struct {
		Origin   string
		Replacer *strings.Replacer

		Options RedirectOptions
	}

	RedirectOptions struct {
		StatusCodeRedirect int
		Schema             string
		KeepURI            bool
	}
)

func NewRedirect(v *viper.Viper) Redirect {
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
		Origin:   v.GetString("redirect.origin"),
		Replacer: r,
		Options: RedirectOptions{
			StatusCodeRedirect: sc,
			Schema:             s,
			KeepURI:            v.GetBool("redirect.options.keep_uri"),
		},
	}
}

func (r Redirect) Redirect(req *http.Request) (string, int) {
	sd := strings.TrimSuffix(req.Host, r.Origin)
	t := fmt.Sprintf("%s%s", r.Options.Schema, r.Replacer.Replace(sd))

	if r.Options.KeepURI && req.RequestURI != "/" {
		t = fmt.Sprintf("%s%s", t, req.RequestURI)
	}

	return t, r.Options.StatusCodeRedirect
}
