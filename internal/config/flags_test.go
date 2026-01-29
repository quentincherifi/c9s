// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/quentincherifi/c9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewFlags(t *testing.T) {
	config.AppDumpsDir = "/tmp/c9s-test/screen-dumps"
	config.AppLogFile = "/tmp/c9s-test/c9s.log"

	f := config.NewFlags()
	assert.InDelta(t, 2.0, *f.RefreshRate, 0.001)
	assert.Equal(t, "info", *f.LogLevel)
	assert.Equal(t, "/tmp/c9s-test/c9s.log", *f.LogFile)
	assert.Equal(t, config.AppDumpsDir, *f.ScreenDumpDir)
	assert.Empty(t, *f.Command)
	assert.False(t, *f.Headless)
	assert.False(t, *f.Logoless)
	assert.False(t, *f.AllNamespaces)
	assert.False(t, *f.ReadOnly)
	assert.False(t, *f.Write)
	assert.False(t, *f.Crumbsless)
	assert.False(t, *f.Splashless)
}
