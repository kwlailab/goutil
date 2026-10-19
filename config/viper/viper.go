package viperconfig

import (
	"errors"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var configOnlyFlags = map[string]struct{}{
	"config":           {},
	"config-directory": {},
}

// Init loads config files and enables env lookup.
// Priority at read time: CLI flags (via ApplyChangedFlags) > env > config file > defaults.
func Init(configDirectory string, configFiles []string) error {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	configPaths := []string{".", configDirectory, "../config", "../../config", "../../../config"}

	for _, name := range configFiles {
		if name == "" {
			continue
		}
		sub := viper.New()
		for _, p := range configPaths {
			sub.AddConfigPath(p)
		}
		sub.SetConfigName(name)
		if err := sub.ReadInConfig(); err != nil {
			if errors.As(err, &viper.ConfigFileNotFoundError{}) {
				continue
			}
			return err
		}
		if err := viper.MergeConfigMap(sub.AllSettings()); err != nil {
			return err
		}
	}

	return nil
}

// ApplyChangedFlags promotes explicitly set CLI flags into viper overrides.
func ApplyChangedFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			return
		}
		if _, skip := configOnlyFlags[flag.Name]; skip {
			return
		}
		viper.Set(flag.Name, flag.Value.String())
	})
}
