package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func loadEnvironmentFile(path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil
	}

	file, err := os.Open(trimmed)
	if err != nil {
		return fmt.Errorf("failed to open env file %q: %w", trimmed, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.IndexRune(line, '=')
		if idx <= 0 {
			return fmt.Errorf("invalid line %d in env file %s", lineNo, trimmed)
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key == "" {
			return fmt.Errorf("invalid line %d in env file %s", lineNo, trimmed)
		}

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set env var %s from %s: %w", key, trimmed, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read env file %s: %w", trimmed, err)
	}

	return nil
}

func loadConfigFile(path string) error {
	viper.Reset()

	viper.SetEnvPrefix("knxctl2")
	replacer := strings.NewReplacer(".", "_", "-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if err := bindEnvKeys(
		"server",
		"port",
		"listen.group_file",
		"listen.database_url",
		"serve.listen",
		"serve.event_buffer",
		"serve.log_events",
		"serve.group_file",
		"serve.response_timeout",
		"serve.database_url",
		"send.group",
		"send.group_name",
		"send.dpt",
		"send.value",
		"send.group_file",
		"send.wait_response",
		"send.timeout",
		"bridge.other",
	); err != nil {
		return err
	}

	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil
	}

	viper.SetConfigFile(trimmed)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %q: %w", trimmed, err)
	}

	return nil
}

func bindEnvKeys(keys ...string) error {
	for _, key := range keys {
		if err := viper.BindEnv(key); err != nil {
			return fmt.Errorf("failed to bind env key %s: %w", key, err)
		}
	}
	return nil
}

func applyGlobalConfig(cmd *cobra.Command) {
	if value := strings.TrimSpace(viper.GetString("server")); value != "" && !flagChanged(cmd, "server") {
		server = value
	}
	if value := strings.TrimSpace(viper.GetString("port")); value != "" && !flagChanged(cmd, "port") {
		port = value
	}
}

func flagChanged(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}

	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag.Changed
	}
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		return flag.Changed
	}
	return false
}

func normalizeDuration(value any) (time.Duration, bool) {
	switch typed := value.(type) {
	case time.Duration:
		return typed, true
	case string:
		parsed, err := time.ParseDuration(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return parsed, true
	case int:
		return time.Duration(typed), true
	case int64:
		return time.Duration(typed), true
	case float64:
		return time.Duration(typed), true
	default:
		return 0, false
	}
}
