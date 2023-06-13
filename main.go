package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/webdevops/kube-pool-manager/config"
	"github.com/webdevops/kube-pool-manager/manager"
)

const (
	Author = "webdevops.io"
)

var (
	argparser *flags.Parser

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

var opts config.Opts

func main() {
	initArgparser()
	initLogger()

	logger.Infof("starting kube-pool-manager v%s (%s; %s; by %v)", gitTag, gitCommit, runtime.Version(), Author)
	logger.Info(string(opts.GetJson()))

	poolManager := manager.KubePoolManager{
		Opts:   opts,
		Config: parseAppConfig(opts.Config),
		Logger: logger,
	}
	poolManager.Init()
	poolManager.Start()

	logger.Infof("starting http server on %s", opts.Server.Bind)
	startHttpServer()
}

func initArgparser() {
	argparser = flags.NewParser(&opts, flags.Default)
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

	logger.With(zap.String("path", path)).Infof("reading configuration from file %v", path)
	/* #nosec */
	if data, err := os.ReadFile(path); err == nil {
		configRaw = data
	} else {
		logger.Fatal(err)
	}

	logger.With(zap.String("path", path)).Info("parsing configuration")
	if err := yaml.Unmarshal(configRaw, &conf); err != nil {
		logger.Fatal(err)
	}

	return
}

func startHttpServer() {
	mux := http.NewServeMux()

	// healthz
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err)
		}
	})

	// readyz
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err)
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         opts.Server.Bind,
		Handler:      mux,
		ReadTimeout:  opts.Server.ReadTimeout,
		WriteTimeout: opts.Server.WriteTimeout,
	}
	logger.Fatal(srv.ListenAndServe())
}
