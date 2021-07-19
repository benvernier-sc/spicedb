package main

import (
	"errors"
	"net/http"
	"net/http/pprof"

	"github.com/jzelinskie/cobrautil"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/authzed/spicedb/pkg/cmdutil"
)

func main() {
	rootCmd := newRootCmd()

	cmdutil.RegisterLoggingPersistentFlags(rootCmd)
	cmdutil.RegisterTracingPersistentFlags(rootCmd)

	registerMigrateCmd(rootCmd)
	registerHeadCmd(rootCmd)
	registerDeveloperServiceCmd(rootCmd)

	rootCmd.Execute()
}

func NewTlsGrpcServer(certPath, keyPath string, opts ...grpc.ServerOption) (*grpc.Server, error) {
	if certPath == "" || keyPath == "" {
		return nil, errors.New("missing one of required values: cert path, key path")
	}

	creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	opts = append(opts, grpc.Creds(creds))
	return grpc.NewServer(opts...), nil
}

func NewMetricsServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := cobrautil.SyncViperPreRunE("spicedb")(cmd, args); err != nil {
		return err
	}

	cmdutil.LoggingPreRun(cmd, args)
	cmdutil.TracingPreRun(cmd, args)

	return nil
}
