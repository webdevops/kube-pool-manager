package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	yaml "github.com/goccy/go-yaml"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/webdevops/kube-pool-manager/config"
	"github.com/webdevops/kube-pool-manager/manager"
)

const (
	Author = "webdevops.io"
)

var (
	argparser *flags.Parser
	Opts      config.Opts

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
	buildDate = "<unknown>"
)

func main() {
	initArgparser()
	initLogger()

	logger.Infof("starting kube-pool-manager v%s (%s; %s; by %v at %v)", gitTag, gitCommit, runtime.Version(), Author, buildDate)
	logger.Info(string(Opts.GetJson()))
	initSystem()

	poolManager := manager.KubePoolManager{
		Opts:   Opts,
		Config: parseAppConfig(Opts.Config),
		Logger: logger,
	}
	poolManager.Init()
	poolManager.Start()

	logger.Infof("starting http server on %s", Opts.Server.Bind)
	startHttpServer()
}

func initArgparser() {
	argparser = flags.NewParser(&Opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

func parseAppConfig(path string) (conf config.Config) {
	var configRaw []byte

	conf = config.Config{}

	logger.With(slog.String("path", path)).Infof("reading configuration from file %v", path)
	/* #nosec */
	configRaw, err := os.ReadFile(path)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.With(slog.String("path", path)).Info("parsing configuration")
	err = yaml.UnmarshalWithOptions(configRaw, &conf, yaml.Strict(), yaml.UseJSONUnmarshaler())
	if err != nil {
		logger.Fatal(err.Error())
	}

	return
}

func startHttpServer() {
	mux := http.NewServeMux()

	// healthz
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err.Error())
		}
	})

	// readyz
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err.Error())
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         Opts.Server.Bind,
		Handler:      mux,
		ReadTimeout:  Opts.Server.ReadTimeout,
		WriteTimeout: Opts.Server.WriteTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal(err.Error())
	}
}
