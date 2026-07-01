package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/spf13/cobra"
	"github.com/yylego/done"
	"github.com/yylego/kratos-examples/demo1kratos/cmd/demo1kratos/cfgpath"
	"github.com/yylego/kratos-examples/demo1kratos/cmd/demo1kratos/subcmds"
	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/appcfg"
	"github.com/yylego/must"
	"github.com/yylego/must/mustslice"
	"github.com/yylego/rese"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
)

func init() {
	fmt.Println("service-name:", Name)
}

func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(done.VCE(os.Hostname()).Omit()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	logger := log.NewLogger(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		}),
		log.WithExtractor(tracing.TraceAttrs),
	).With(
		slog.String("service.id", done.VCE(os.Hostname()).Omit()),
		slog.String("service.name", Name),
		slog.String("service.version", Version),
	)
	log.SetDefault(logger)

	var rootCmd = &cobra.Command{
		Use:   "demo1kratos",
		Short: "A Kratos microservice with database migration",
		Run: func(cmd *cobra.Command, args []string) {
			mustslice.None(args)
			if cfg := appcfg.ParseConfig(cfgpath.ConfigPath); cfg.Server.AutoRun {
				runApp(cfg, logger)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "conf", "c", "./configs", "config path, eg: --conf=config.yaml")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Start the application",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
			runApp(cfg, logger)
		},
	})

	rootCmd.AddCommand(subcmds.NewVersionCmd(Name, Version, logger))
	rootCmd.AddCommand(subcmds.NewMigrateCmd(logger))

	must.Done(rootCmd.Execute())
}

func runApp(cfg *conf.Bootstrap, logger *slog.Logger) {
	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
	defer cleanup()

	// start and wait for stop signal
	must.Done(app.Run())
}
