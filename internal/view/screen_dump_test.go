// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/quentincherifi/c9s/internal/client"
	"github.com/quentincherifi/c9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScreenDumpNew(t *testing.T) {
	po := view.NewScreenDump(client.SdGVR)

	require.NoError(t, po.Init(makeCtx(t)))
	assert.Equal(t, "ScreenDumps", po.Name())
	assert.Len(t, po.Hints(), 7)
}
