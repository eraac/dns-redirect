package main

import (
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewRedirect(t *testing.T) {
	v := NewViper(config{
		"resolver.type": "dns",
		"redirect.host": ".dev.local.",
		"redirect.options.keyword.slash": "--slash--",
		"redirect.options.keyword.dot": "--dot--",
		"redirect.options.keyword.colon": "--colon--",
		"redirect.options.keyword.interrogation-mark": "--int--",
		"redirect.options.keyword.ampersand": "--amp--",
		"redirect.options.keyword.equal": "--equal--",
		"redirect.options.keyword.percent": "--percent--",
		"redirect.options.permanent_redirect": true,
		"redirect.options.enforce_https": false,
		"redirect.options.keep_uri": true,
	})

	r, err := NewRedirect(logrus.StandardLogger(), v)
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}

	if r.host != ".dev.local." {
		t.Errorf("expected %s, got %s", ".dev.local.", r.host)
	}

	if r.options.statusCodeRedirect != http.StatusPermanentRedirect {
		t.Errorf("expected %d, got %d", http.StatusPermanentRedirect, r.options.statusCodeRedirect)
	}

	if r.options.schema != "" {
		t.Errorf("expected \"\", got %s", r.options.schema)
	}

	if !r.options.keepURI {
		t.Error("expected true, got false")
	}
}

func TestNewRedirect_config(t *testing.T) {
	v := NewViper(config{
		"resolver.type": "static",
		"redirect.host": ".dev.local.",
		"redirect.options.keyword.slash": "--slash--",
		"redirect.options.keyword.dot": "--dot--",
		"redirect.options.keyword.colon": "--colon--",
		"redirect.options.keyword.interrogation-mark": "--int--",
		"redirect.options.keyword.ampersand": "--amp--",
		"redirect.options.keyword.equal": "--equal--",
		"redirect.options.keyword.percent": "--percent--",
		"redirect.options.permanent_redirect": false,
		"redirect.options.enforce_https": true,
		"redirect.options.keep_uri": false,
	})

	r, err := NewRedirect(logrus.StandardLogger(), v)
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}

	if r.host != ".dev.local." {
		t.Errorf("expected %s, got %s", ".dev.local.", r.host)
	}

	if r.options.statusCodeRedirect != http.StatusTemporaryRedirect {
		t.Errorf("expected %d, got %d", http.StatusTemporaryRedirect, r.options.statusCodeRedirect)
	}

	if r.options.schema != "https://" {
		t.Errorf("expected \"https://\", got %s", r.options.schema)
	}

	if r.options.keepURI {
		t.Error("expected false, got true")
	}
}

func TestNewRedirect_err(t *testing.T) {
	v := NewViper(config{
		"resolver.type": "nope",
		"redirect.host": ".dev.local.",
	})

	r, err := NewRedirect(logrus.StandardLogger(), v)
	if err != UnknownResolverTypeErr {
		t.Errorf("expected UnknowResolverTypeErr, got %s", err)
	}

	if r != nil {
		t.Errorf("expected nil, got %T", r)
	}
}

func TestNewRedirect_Redirect(t *testing.T) {
	v := NewViper(config{
		"resolver.type": "dns",
		"redirect.host": ".local.labesse.me.",
		"redirect.options.keyword.slash": "--slash--",
		"redirect.options.keyword.dot": "--dot--",
		"redirect.options.keyword.colon": "--colon--",
		"redirect.options.keyword.interrogation-mark": "--int--",
		"redirect.options.keyword.ampersand": "--amp--",
		"redirect.options.keyword.equal": "--equal--",
		"redirect.options.keyword.percent": "--percent--",
		"redirect.options.permanent_redirect": true,
		"redirect.options.enforce_https": false,
		"redirect.options.keep_uri": true,
	})

	r, err := NewRedirect(logrus.StandardLogger(), v)
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}

	inputs := []struct{
		in *http.Request
		outHost string
		outRedirectCode int
		errExpected bool
	}{
		{in: &http.Request{Host: "dns-redirect.labesse.me"}, outHost: "google.com", outRedirectCode: http.StatusPermanentRedirect, errExpected: false},
		{in: &http.Request{Host: "dns-redirect.labesse.me", RequestURI: "/uri"}, outHost: "google.com/uri", outRedirectCode: http.StatusPermanentRedirect, errExpected: false},
		{in: &http.Request{Host: "invalid-host"}, outHost: "", outRedirectCode: 0, errExpected: true},
	}

	for _, i := range inputs {
		h, c, err := r.Redirect(i.in)

		if h != i.outHost {
			t.Errorf("expected %s, got %s", i.outHost, h)
		}

		if c != i.outRedirectCode {
			t.Errorf("expected %d, got %d", i.outRedirectCode, c)
		}

		if err == nil && i.errExpected {
			t.Errorf("expected err, got nil")
		}

		if err != nil && !i.errExpected {
			t.Errorf("expected nil, get %s", err)
		}
	}
}