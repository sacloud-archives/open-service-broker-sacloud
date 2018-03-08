package main

import (
	"fmt"

	"gopkg.in/urfave/cli.v2"
	"strings"
)

type cliConfig struct {
	AccessToken       string
	AccessTokenSecret string
	Zone              string
	AcceptLanguage    string
	RetryMax          int
	RetryIntervalSec  int64
	APIRootURL        string
	TraceMode         bool
	BasicAuthUsername string
	BasicAuthPassword string
	LogLevel          string
}

var cfg = &cliConfig{}

// DefaultZone is used when zone is unspecified
var DefaultZone = "tk1a"

var cliFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "token",
		Usage:       "API Token of SakuraCloud",
		EnvVars:     []string{"SAKURACLOUD_ACCESS_TOKEN"},
		Destination: &cfg.AccessToken,
	},
	&cli.StringFlag{
		Name:        "secret",
		Usage:       "API Secret of SakuraCloud",
		EnvVars:     []string{"SAKURACLOUD_ACCESS_TOKEN_SECRET"},
		Destination: &cfg.AccessTokenSecret,
	},
	&cli.StringFlag{
		Name:        "zone",
		Usage:       "Target zone of SakuraCloud",
		EnvVars:     []string{"SAKURACLOUD_ZONE"},
		Value:       DefaultZone,
		Destination: &cfg.Zone,
	},
	&cli.StringFlag{
		Name:        "accept-language",
		Usage:       "Accept-Language Header",
		EnvVars:     []string{"SAKURACLOUD_ACCEPT_LANGUAGE"},
		Destination: &cfg.AcceptLanguage,
	},
	&cli.IntFlag{
		Name:        "retry-max",
		Usage:       "Number of API-Client retries",
		EnvVars:     []string{"SAKURACLOUD_RETRY_MAX"},
		Destination: &cfg.RetryMax,
	},
	&cli.Int64Flag{
		Name:        "retry-interval",
		Usage:       "API client retry interval seconds",
		EnvVars:     []string{"SAKURACLOUD_RETRY_INTERVAL"},
		Destination: &cfg.RetryIntervalSec,
	},
	&cli.StringFlag{
		Name:        "api-root-url",
		EnvVars:     []string{"SAKURACLOUD_API_ROOT_URL"},
		Destination: &cfg.APIRootURL,
		Hidden:      true,
	},
	&cli.BoolFlag{
		Name:        "trace",
		Usage:       "Flag of SakuraCloud debug-mode",
		EnvVars:     []string{"SAKURACLOUD_TRACE_MODE"},
		Destination: &cfg.TraceMode,
		Value:       false,
		Hidden:      true,
	},
	&cli.StringFlag{
		Name:        "basic-auth-username",
		Usage:       "BASIC auth username",
		EnvVars:     []string{"BASIC_AUTH_USERNAME"},
		Destination: &cfg.BasicAuthUsername,
	},
	&cli.StringFlag{
		Name:        "basic-auth-password",
		Usage:       "BASIC auth password",
		EnvVars:     []string{"BASIC_AUTH_PASSWORD"},
		Destination: &cfg.BasicAuthPassword,
	},
	&cli.StringFlag{
		Name:        "log-level",
		Usage:       "Log level[INFO/WARN/DEBUG] default:INFO",
		EnvVars:     []string{"OSBS_LOG_LEVEL"},
		Value:       "INFO",
		Destination: &cfg.LogLevel,
	},
}

func (o *cliConfig) Validate() []error {
	var errs []error

	validators := []func() error{
		func() error { return o.validateRequired("token", o.AccessToken) },
		func() error { return o.validateRequired("secret", o.AccessTokenSecret) },
		func() error { return o.validateRequired("zone", o.Zone) },
		func() error { return o.validateRequired("log-level", o.LogLevel) },
		func() error { return o.validateInStrings("log-level", o.LogLevel, "INFO", "WARN", "DEBUG") },
	}

	for _, v := range validators {
		err := v()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (o *cliConfig) validateRequired(name, v string) error {
	if v == "" {
		return fmt.Errorf("[Option] --%s is required", name)
	}
	return nil
}

func (o *cliConfig) validateInStrings(name, v string, allows ...string) error {
	if v == "" {
		return nil
	}
	exists := false
	for _, allow := range allows {
		if v == allow {
			exists = true
			break
		}
	}
	if !exists {
		return fmt.Errorf("[Option] --%s must be in [%s]", name, strings.Join(allows, "/"))
	}
	return nil
}
