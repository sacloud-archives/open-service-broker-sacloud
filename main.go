package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sacloud/open-service-broker-sacloud/broker"
	"github.com/sacloud/open-service-broker-sacloud/iaas"
	"github.com/sacloud/open-service-broker-sacloud/service"
	"github.com/sacloud/open-service-broker-sacloud/version"
	"gopkg.in/urfave/cli.v2"
)

var (
	appName      = "open-service-broker-sacloud"
	appUsage     = "An implementation of the Open Service Broker API for SAKURA cloud"
	appCopyright = "Copyright (C) 2018 Kazumichi Yamamoto."
)

func main() {
	app := &cli.App{
		Name:                  appName,
		Usage:                 appUsage,
		HelpName:              appName,
		Copyright:             appCopyright,
		EnableShellCompletion: true,
		Version:               version.FullVersion(),
		CommandNotFound:       cmdNotFound,
		Flags:                 cliFlags,
		Action:                cmdMain,
	}
	cli.InitCompletionFlag.Hidden = true

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func cmdNotFound(c *cli.Context, command string) {
	fmt.Fprintf(
		os.Stderr,
		"%s: '%s' is not a %s command. See '%s --help'\n",
		c.App.Name,
		command,
		c.App.Name,
		c.App.Name,
	)
	os.Exit(1)
}

func cmdMain(c *cli.Context) error {

	errs := cfg.Validate()
	if len(errs) > 0 {
		return flattenErrors(errs...)
	}

	// Initialize log setting
	log.SetOutput(os.Stdout)
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	logLevel := log.InfoLevel
	switch cfg.LogLevel {
	case "WARN":
		logLevel = log.WarnLevel
	case "DEBUG":
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)

	log.WithFields(
		log.Fields{
			"version": version.FullVersion(),
		},
	).Info("Start Open Service Broker for SAKURA Cloud")

	// prepare SAKURA cloud API client
	sacloudAPI := iaas.NewClient(&iaas.ClientConfig{
		AccessToken:       cfg.AccessToken,
		AccessTokenSecret: cfg.AccessTokenSecret,
		Zone:              cfg.Zone,
		AcceptLanguage:    cfg.AcceptLanguage,
		RetryMax:          cfg.RetryMax,
		RetryIntervalSec:  cfg.RetryIntervalSec,
		APIRootURL:        cfg.APIRootURL,
		TraceMode:         cfg.TraceMode,
	})
	err := service.Initialize(sacloudAPI)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-sigChan
		log.WithField(
			"signal",
			signal,
		).Debug("signal received; shutting down")
		log.Info("Shutting down...")
		cancel()
	}()

	// Start broker(s)
	brokerCfg := &broker.Config{
		Port:              8080, // TODO make configurable
		BasicAuthUsername: cfg.BasicAuthUsername,
		BasicAuthPassword: cfg.BasicAuthPassword,
	}
	b := broker.NewBroker(brokerCfg)
	if err := b.Start(ctx); err != nil {
		if err == ctx.Err() {
			time.Sleep(time.Second * 5) // sleep for shutting down goroutines
		} else {
			log.Fatal(err)
		}
	}

	log.Info("Shutdown complete")
	return nil
}

func flattenErrors(errors ...error) error {
	if len(errors) == 0 {
		return nil
	}
	var list = make([]string, 0)
	for _, str := range errors {
		list = append(list, str.Error())
	}
	return fmt.Errorf(strings.Join(list, "\n"))
}
