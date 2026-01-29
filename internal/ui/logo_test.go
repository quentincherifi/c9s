// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/quentincherifi/c9s/internal/config"
	"github.com/quentincherifi/c9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestNewLogoView(t *testing.T) {
	v := ui.NewLogo(config.NewStyles())
	v.Reset()

	const elogo = "[#ffa500::b]  ____ ___  ____       \n[#ffa500::b] / ___|__ \\/ ___|      \n[#ffa500::b]| |    / _/\\___ \\      \n[#ffa500::b]| |___| |   ___) |     \n[#ffa500::b] \\____|_|  |____/      \n[#ffa500::b]  Claude + K9s = C9s   \n"
	assert.Equal(t, elogo, v.Logo().GetText(false))
	assert.Empty(t, v.Status().GetText(false))
}

func TestLogoStatus(t *testing.T) {
	uu := map[string]struct {
		logo, msg, e string
	}{
		"info": {
			"[#008000::b]  ____ ___  ____       \n[#008000::b] / ___|__ \\/ ___|      \n[#008000::b]| |    / _/\\___ \\      \n[#008000::b]| |___| |   ___) |     \n[#008000::b] \\____|_|  |____/      \n[#008000::b]  Claude + K9s = C9s   \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
		"warn": {
			"[#c71585::b]  ____ ___  ____       \n[#c71585::b] / ___|__ \\/ ___|      \n[#c71585::b]| |    / _/\\___ \\      \n[#c71585::b]| |___| |   ___) |     \n[#c71585::b] \\____|_|  |____/      \n[#c71585::b]  Claude + K9s = C9s   \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
		"err": {
			"[#ff0000::b]  ____ ___  ____       \n[#ff0000::b] / ___|__ \\/ ___|      \n[#ff0000::b]| |    / _/\\___ \\      \n[#ff0000::b]| |___| |   ___) |     \n[#ff0000::b] \\____|_|  |____/      \n[#ff0000::b]  Claude + K9s = C9s   \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
	}

	v := ui.NewLogo(config.NewStyles())
	for n := range uu {
		k, u := n, uu[n]
		t.Run(k, func(t *testing.T) {
			switch k {
			case "info":
				v.Info(u.msg)
			case "warn":
				v.Warn(u.msg)
			case "err":
				v.Err(u.msg)
			}
			assert.Equal(t, u.logo, v.Logo().GetText(false))
			assert.Equal(t, u.e, v.Status().GetText(false))
		})
	}
}
