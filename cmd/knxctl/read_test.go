package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestApplyReadConfigDefaultWaitResponse(t *testing.T) {
	originalWait := waitForResponse
	t.Cleanup(func() {
		waitForResponse = originalWait
	})

	viper.Reset()
	t.Cleanup(viper.Reset)

	cmd := &cobra.Command{}
	cmd.Flags().BoolVar(&waitForResponse, "wait-response", true, "")

	waitForResponse = false

	applyReadConfig(cmd)

	if !waitForResponse {
		t.Fatalf("expected waitForResponse to default to true when unset")
	}
}
