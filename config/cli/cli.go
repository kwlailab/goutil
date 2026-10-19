package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"

	"github.com/kwlailab/goutil/config/viper"
	"github.com/kwlailab/goutil/pkg/log"
	"github.com/kwlailab/goutil/pkg/otel"
	"github.com/kwlailab/goutil/pkg/shutdown"
)

// NewRootCmd builds a cobra root: viper config, optional tracer, then log.ReplaceGlobals.
func NewRootCmd(serviceName, version, stage string, defaultConfigFiles ...string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   serviceName,
		Short: serviceName,
	}
	defaults := append([]string{"config", "config.local"}, defaultConfigFiles...)
	configFile := []string{}
	configDirectory := ""
	addr := ""
	cmd.PersistentFlags().StringSliceVarP(&configFile, "config", "c", defaults, "config file name without extension")
	cmd.PersistentFlags().StringVarP(&configDirectory, "config-directory", "d", "./config", "config file directory")
	cmd.PersistentFlags().StringVarP(&addr, "addr", "a", "0.0.0.0:50051", "address to listen on")
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := viperconfig.Init(configDirectory, configFile); err != nil {
			return err
		}
		viperconfig.ApplyChangedFlags(cmd.PersistentFlags())

		viper.SetDefault("version", version)
		viper.SetDefault("stage", stage)
		viper.SetDefault("service", serviceName)

		fmt.Println("Configuration initialized. Using config file:", viper.ConfigFileUsed())

		otelEndpoint := viper.GetString("otel.exporter.otlp.endpoint")
		if otelEndpoint == "" {
			otelEndpoint = "localhost:4317"
		}
		if !viper.GetBool("otel.disabled") {
			attributes := []attribute.KeyValue{
				semconv.ServiceName(serviceName),
				semconv.ServiceVersion(version),
				semconv.DeploymentEnvironment(viper.GetString("env")),
				attribute.String("stage", stage),
			}
			shutdownTracer, err := otel.InitTracer(context.Background(), otelEndpoint, attributes...)
			if err != nil {
				fmt.Println("Failed to initialize tracer: skip!", err)
			} else if shutdownTracer != nil {
				shutdown.Register(shutdownTracer)
			}
		}

		zapProduction := viper.GetBool("log.zap.production")
		var zapLogger *zap.Logger
		if zapProduction {
			zapLogger, _ = zap.NewProduction()
		} else {
			zapLogger, _ = zap.NewDevelopment()
		}
		log.ReplaceGlobals(otelzap.New(zapLogger))
		fmt.Println("Logger initialized. zap log production:", zapProduction)
		return nil
	}
	return cmd
}

func BuildInfo() map[string]any {
	inst, _ := os.Hostname()
	return map[string]any{
		"now":     time.Now(),
		"inst":    inst,
		"version": viper.GetString("version"),
		"build":   viper.GetString("build"),
		"stage":   viper.GetString("stage"),
	}
}
