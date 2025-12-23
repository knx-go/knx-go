package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	server          string
	port            string
	bridgeOther     string
	group           string
	groupName       string
	valueRaw        string
	writeDPT        string
	waitForResponse bool
	waitTimeout     time.Duration = 5 * time.Second
	groupFile       string
	postgresqlDSN   string
	configPath      string
	envFile         string
)

var root = &cobra.Command{
	Use:   "knxctl2",
	Short: "KNX multi-tool with listening, bridging, sending, and HTTP serving capabilities",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(envFile) == "" {
			envFile = os.Getenv("KNXCTL2_ENV_FILE")
		}
		if err := loadEnvironmentFile(envFile); err != nil {
			return err
		}

		if strings.TrimSpace(configPath) == "" {
			if value := os.Getenv("KNXCTL2_CONFIG"); value != "" {
				configPath = value
			}
		}

		if err := loadConfigFile(configPath); err != nil {
			return err
		}

		applyGlobalConfig(cmd)
		return nil
	},
}

func init() {
	root.PersistentFlags().StringVarP(&envFile, "env-file", "E", "", "path to an environment file containing KEY=VALUE pairs")
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "", "path to an INI/TOML configuration file")
	root.PersistentFlags().StringVarP(&server, "server", "s", "127.0.0.1", "KNXnet/IP server address")
	root.PersistentFlags().StringVarP(&port, "port", "p", "3671", "KNXnet/IP server port")
}

func main() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
