package main

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/sirupsen/logrus"
)

func TestNewResolverFromConfig(t *testing.T) {
	vDNS := NewViper(config{"resolver.type": "dns"})
	r, err := NewResolverFromConfig(logrus.StandardLogger(), vDNS)
	if err != nil {
		t.Errorf("expected nil, got %s", err)
	}

	if _, ok := r.(*DNSResolver); !ok {
		t.Errorf("expected type *DNSResolver, got %T", r)
	}

	vStatic := NewViper(config{"resolver.type": "static"})
	r, err = NewResolverFromConfig(logrus.StandardLogger(), vStatic)
	if err != nil {
		t.Errorf("expected nil, got %s", err)
	}

	if _, ok := r.(*StaticResolver); !ok {
		t.Errorf("expected type *StaticResolver, got %T", r)
	}

	vFail := NewViper(config{"resolver.type": "nope"})
	r, err = NewResolverFromConfig(logrus.StandardLogger(), vFail)
	if err != UnknownResolverTypeErr {
		t.Errorf("expected UnknowResolverTypeErr, got %s", err)
	}

	if r != nil {
		t.Errorf("expected nil, got %T", r)
	}
}

func TestNewDNSResolver(t *testing.T) {
	v := NewViper(config{"resolver.config.dns_server": "1.1.1.1:53", "resolver.config.timeout": 10})

	r, err := NewDNSResolver(logrus.StandardLogger(), v)
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}

	d, ok := r.(*DNSResolver)
	if !ok {
		t.Fatalf("expected type *DNSResolver, got %T", r)
	}

	if !d.resolver.PreferGo {
		t.Error("expected true, got false")
	}
}

func TestDNSResolver_Resolve(t *testing.T) {
	v := NewViper(config{"resolver.config.dns_server": "1.1.1.1:53", "resolver.config.timeout": 10})

	r, err := NewDNSResolver(logrus.StandardLogger(), v)
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}

	inputs := []struct{
		in string
		out string
		errExpected bool
	}{
		{in: "dns-redirect.labesse.me", out: "google--dot--com.local.labesse.me.", errExpected: false},
		{in: "toto.local", out: "", errExpected: true},
	}

	for _, i := range inputs {
		out, err := r.Resolve(i.in)

		if out != i.out {
			t.Errorf("expected %s, got %s", i.out, out)
		}

		if err == nil && i.errExpected {
			t.Errorf("expecterd err, got nil")
		}

		if err != nil && !i.errExpected {
			t.Errorf("expecterd nil, got %s", err)
		}
	}
}

func TestNewStaticResolver(t *testing.T) {
	hosts := map[string]string{
		"dev.local": "google--dot--com.dev.local.",
		"staging.local": "labesse--dot--me.dev.local.",
	}

	v := NewViper(config{
		"resolver.config.hosts": hosts,
	})

	r, err := NewStaticResolver(logrus.StandardLogger(), v)
	if err != nil {
		t.Fatalf("expected <nil>, got %s", err)
	}

	sr, ok := r.(*StaticResolver)
	if !ok {
		t.Fatalf("expected type *StaticResolver, got %T", r)
	}

	if diff := deep.Equal(sr.hosts, hosts); diff != nil {
		t.Error(diff)
	}
}

func TestStaticResolver_Resolve(t *testing.T) {
	hosts := map[string]string{
		"dev.local": "google--dot--com.dev.local.",
		"staging.local": "labesse--dot--me.dev.local.",
	}

	v := NewViper(config{
		"resolver.config.hosts": hosts,
	})

	sr, err := NewStaticResolver(logrus.StandardLogger(), v)
	if err != nil {
		t.Fatalf("expected <nil>, got %s", err)
	}

	inputs := []struct{
		in string
		out string
		errExpected error
	}{
		{in: "dev.local", out: "google--dot--com.dev.local.", errExpected: nil},
		{in: "toto.local", out: "", errExpected: HostNotFoundErr},
		{in: "staging.local", out: "labesse--dot--me.dev.local.", errExpected: nil},
	}

	for _, i := range inputs {
		out, err := sr.Resolve(i.in)

		if out != i.out {
			t.Errorf("expected %s, got %s", i.out, out)
		}

		if err != i.errExpected {
			t.Errorf("expecterd err %s, got error %s", i.errExpected, err)
		}
	}
}