package main

import (
	"context"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	DNSResolverType    = "dns"
	StaticResolverType = "static"

	UnknownResolverTypeErr Error = "unknown resolver type"
	HostNotFoundErr Error = "host not found"
)

type (
	Resolver interface {
		Resolve(string) (string, error)
	}

	DNSResolver struct {
		logger   logrus.FieldLogger
		resolver net.Resolver
	}

	StaticResolver struct {
		logger logrus.FieldLogger
		hosts  map[string]string
	}
)

func NewResolverFromConfig(l logrus.FieldLogger, v *viper.Viper) (Resolver, error) {
	t := v.GetString("resolver.type")
	l.WithField("resolver_type", t).Info("init resolver")

	switch t {
	case DNSResolverType:
		return NewDNSResolver(l, v)
	case StaticResolverType:
		return NewStaticResolver(l, v)
	}

	return nil, UnknownResolverTypeErr
}

func NewDNSResolver(l logrus.FieldLogger, v *viper.Viper) (Resolver, error) {
	r := &DNSResolver{
		logger:   l,
		resolver: net.Resolver{},
	}

	if s := v.GetString("resolver.config.dns_server"); s != "" {
		r.resolver.PreferGo = true

		timeout := time.Duration(v.GetInt("resolver.config.timeout")) * time.Second

		r.resolver.Dial = func(ctx context.Context, network, _ string) (conn net.Conn, e error) {
			d := net.Dialer{Timeout: timeout}
			return d.DialContext(ctx, network, s)
		}
	}

	return r, nil
}

func NewStaticResolver(l logrus.FieldLogger, v *viper.Viper) (Resolver, error) {
	return &StaticResolver{
		logger: l,
		hosts: v.GetStringMapString("resolver.config.hosts"),
	}, nil
}

func (r *DNSResolver) Resolve(h string) (string, error) {
	r.logger.WithField("host", h).Debug("resolve")

	return r.resolver.LookupCNAME(context.Background(), h)
}

func (r *StaticResolver) Resolve(h string) (string, error) {
	r.logger.WithField("host", h).Debug("resolve")

	res, ok := r.hosts[h]
	if !ok {
		return "", HostNotFoundErr
	}

	return res, nil
}