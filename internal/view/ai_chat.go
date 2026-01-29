// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/ai"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	aiTitle    = "AI Assistant"
	aiTitleFmt = " [aqua::b]%s [gray](%s) "
)

// AIChat represents the AI chat view.
type AIChat struct {
	*tview.Flex

	app          *App
	actions      *ui.KeyActions
	cmdBuff      *model.FishBuff
	chatHistory  *tview.TextView
	contextInfo  *tview.TextView
	messages     []ai.Message
	k8sContext   *ai.K8sContext
	provider     ai.Provider
	providerName string
}

// NewAIChat returns a new AI chat view instance.
func NewAIChat(app *App, question string) *AIChat {
	c := &AIChat{
		Flex:        tview.NewFlex(),
		app:         app,
		actions:     ui.NewKeyActions(),
		cmdBuff:     model.NewFishBuff(':', model.CommandBuffer),
		chatHistory: tview.NewTextView(),
		contextInfo: tview.NewTextView(),
		messages:    make([]ai.Message, 0),
	}

	c.buildK8sContext()

	if question != "" {
		c.messages = append(c.messages, ai.Message{
			Role:    "user",
			Content: question,
		})
	}

	return c
}

func (*AIChat) SetCommand(*cmd.Interpreter)            {}
func (*AIChat) SetFilter(string, bool)                 {}
func (*AIChat) SetLabelSelector(labels.Selector, bool) {}

// Init initializes the view.
func (c *AIChat) Init(_ context.Context) error {
	// Initialize provider
	var err error
	c.provider, err = c.app.Config.K9s.AI.CreateProvider()
	if err != nil {
		c.providerName = "Error"
	} else {
		c.providerName = c.provider.Name()
	}

	c.SetDirection(tview.FlexRow)
	c.SetBorder(true)
	c.SetTitle(fmt.Sprintf(aiTitleFmt, aiTitle, c.providerName))
	c.SetTitleColor(tcell.ColorAqua)
	c.SetBorderPadding(0, 0, 1, 1)

	// Context info section
	c.contextInfo.SetDynamicColors(true)
	c.contextInfo.SetBorder(true)
	c.contextInfo.SetTitle(" Context ")
	c.contextInfo.SetBorderColor(tcell.ColorGray)
	c.updateContextDisplay()

	// Chat history section
	c.chatHistory.SetDynamicColors(true)
	c.chatHistory.SetScrollable(true)
	c.chatHistory.SetWrap(true)
	c.chatHistory.SetBorder(true)
	c.chatHistory.SetTitle(" Chat ")
	c.chatHistory.SetBorderColor(tcell.ColorGray)

	// Layout
	c.AddItem(c.contextInfo, 5, 0, false)
	c.AddItem(c.chatHistory, 0, 1, true)

	c.app.Styles.AddListener(c)
	c.StylesChanged(c.app.Styles)

	c.bindKeys()
	c.SetInputCapture(c.keyboard)

	c.app.Prompt().SetModel(c.cmdBuff)
	c.cmdBuff.AddListener(c)

	// If we have an initial question, send it
	if len(c.messages) > 0 {
		go c.sendMessage()
	}

	return nil
}

func (c *AIChat) buildK8sContext() {
	c.k8sContext = &ai.K8sContext{}

	if c.app.Conn() != nil && c.app.Conn().ConnectionOK() {
		cfg := c.app.Conn().Config()
		if cfg != nil {
			c.k8sContext.ContextName = c.app.Config.ActiveContextName()
			if clusterName, err := cfg.CurrentClusterName(); err == nil {
				c.k8sContext.ClusterName = clusterName
			}
		}
	}

	c.k8sContext.Namespace = c.app.Config.ActiveNamespace()

	// Get current view info
	if top := c.app.Content.Top(); top != nil {
		c.k8sContext.ResourceType = top.Name()
	}

	// Try to get selected resource info from the current view
	c.extractSelectedResource()
}

func (c *AIChat) extractSelectedResource() {
	top := c.app.Content.Top()
	if top == nil {
		return
	}

	// Try to get ResourceViewer interface
	if rv, ok := top.(ResourceViewer); ok {
		if tbl := rv.GetTable(); tbl != nil {
			if sel := tbl.GetSelectedItem(); sel != "" {
				c.k8sContext.SelectedResource = sel
			}
		}
	}
}

func (c *AIChat) updateContextDisplay() {
	var sb strings.Builder
	sb.WriteString("[yellow]Cluster:[white] ")
	if c.k8sContext.ClusterName != "" {
		sb.WriteString(c.k8sContext.ClusterName)
	} else {
		sb.WriteString("N/A")
	}
	sb.WriteString("  [yellow]Context:[white] ")
	if c.k8sContext.ContextName != "" {
		sb.WriteString(c.k8sContext.ContextName)
	} else {
		sb.WriteString("N/A")
	}
	sb.WriteString("  [yellow]Namespace:[white] ")
	if c.k8sContext.Namespace != "" {
		sb.WriteString(c.k8sContext.Namespace)
	} else {
		sb.WriteString("all")
	}
	if c.k8sContext.ResourceType != "" {
		sb.WriteString("  [yellow]View:[white] ")
		sb.WriteString(c.k8sContext.ResourceType)
	}
	if c.k8sContext.SelectedResource != "" {
		sb.WriteString("\n[yellow]Selected:[white] ")
		sb.WriteString(c.k8sContext.SelectedResource)
	}

	c.contextInfo.SetText(sb.String())
}

func (c *AIChat) updateChatDisplay() {
	var sb strings.Builder

	for _, msg := range c.messages {
		switch msg.Role {
		case "user":
			sb.WriteString("[aqua::b]You:[white:-:-] ")
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString(fmt.Sprintf("[green::b]%s:[white:-:-] ", c.providerName))
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		}
	}

	c.chatHistory.SetText(sb.String())
	c.chatHistory.ScrollToEnd()
}

func (c *AIChat) sendMessage() {
	if c.provider == nil {
		c.app.QueueUpdateDraw(func() {
			c.messages = append(c.messages, ai.Message{
				Role:    "assistant",
				Content: "[red]Error: Failed to initialize AI provider. Check your configuration.",
			})
			c.updateChatDisplay()
		})
		return
	}

	// Check API key for providers that need it
	providerType := c.app.Config.K9s.AI.GetProvider()
	if providerType != ai.ProviderOllama {
		apiKey := c.app.Config.K9s.AI.GetAPIKey()
		if apiKey == "" {
			c.app.QueueUpdateDraw(func() {
				var hint string
				switch providerType {
				case ai.ProviderOpenAI:
					hint = "Use ':ai set-key <key>' or set OPENAI_API_KEY"
				default:
					hint = "Use ':ai set-key <key>' or set ANTHROPIC_API_KEY"
				}
				c.messages = append(c.messages, ai.Message{
					Role:    "assistant",
					Content: fmt.Sprintf("[red]Error: API key not configured. %s", hint),
				})
				c.updateChatDisplay()
			})
			return
		}
	}

	// Show loading indicator
	c.app.QueueUpdateDraw(func() {
		c.chatHistory.SetText(c.chatHistory.GetText(false) + "[gray]Thinking...[white]\n")
	})

	systemPrompt, err := ai.BuildSystemPrompt(c.k8sContext)
	if err != nil {
		c.app.QueueUpdateDraw(func() {
			c.messages = append(c.messages, ai.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("[red]Error building prompt: %v", err),
			})
			c.updateChatDisplay()
		})
		return
	}

	resp, err := c.provider.Send(systemPrompt, c.messages)
	if err != nil {
		c.app.QueueUpdateDraw(func() {
			c.messages = append(c.messages, ai.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("[red]Error: %v", err),
			})
			c.updateChatDisplay()
		})
		return
	}

	c.app.QueueUpdateDraw(func() {
		c.messages = append(c.messages, ai.Message{
			Role:    "assistant",
			Content: resp.Content,
		})
		c.updateChatDisplay()
	})
}

func (c *AIChat) bindKeys() {
	c.actions.Bulk(ui.KeyMap{
		tcell.KeyEscape: ui.NewKeyAction("Back", c.backCmd, true),
		ui.KeyQ:         ui.NewKeyAction("Back", c.backCmd, false),
		tcell.KeyEnter:  ui.NewKeyAction("Send", c.sendCmd, true),
		tcell.KeyCtrlL:  ui.NewKeyAction("Clear", c.clearCmd, true),
		ui.KeyColon:     ui.NewSharedKeyAction("Prompt", c.activateCmd, false),
	})
}

func (c *AIChat) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if a, ok := c.actions.Get(ui.AsKey(evt)); ok {
		return a.Action(evt)
	}

	return evt
}

func (c *AIChat) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if c.cmdBuff.InCmdMode() {
		c.cmdBuff.SetActive(false)
		c.cmdBuff.Reset()
		return nil
	}
	return c.app.PrevCmd(evt)
}

func (c *AIChat) sendCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !c.cmdBuff.InCmdMode() || c.cmdBuff.Empty() {
		return evt
	}

	question := c.cmdBuff.GetText()
	c.cmdBuff.SetActive(false)
	c.cmdBuff.Reset()

	c.messages = append(c.messages, ai.Message{
		Role:    "user",
		Content: question,
	})
	c.updateChatDisplay()

	go c.sendMessage()

	return nil
}

func (c *AIChat) clearCmd(*tcell.EventKey) *tcell.EventKey {
	c.messages = make([]ai.Message, 0)
	c.chatHistory.SetText("")
	return nil
}

func (c *AIChat) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if c.app.InCmdMode() {
		return evt
	}
	c.app.ResetPrompt(c.cmdBuff)
	return nil
}

// BufferChanged indicates the buffer was changed.
func (*AIChat) BufferChanged(_, _ string) {}

// BufferCompleted indicates input was accepted.
func (c *AIChat) BufferCompleted(text, _ string) {
	if text == "" {
		return
	}

	c.messages = append(c.messages, ai.Message{
		Role:    "user",
		Content: text,
	})
	c.updateChatDisplay()

	go c.sendMessage()
}

// BufferActive indicates the buff activity changed.
func (c *AIChat) BufferActive(state bool, k model.BufferKind) {
	c.app.BufferActive(state, k)
}

// InCmdMode checks if prompt is active.
func (c *AIChat) InCmdMode() bool {
	return c.cmdBuff.InCmdMode()
}

// StylesChanged notifies the skin changed.
func (c *AIChat) StylesChanged(s *config.Styles) {
	c.SetBackgroundColor(s.BgColor())
	c.chatHistory.SetBackgroundColor(s.BgColor())
	c.chatHistory.SetTextColor(s.FgColor())
	c.contextInfo.SetBackgroundColor(s.BgColor())
	c.contextInfo.SetTextColor(s.FgColor())
	c.SetBorderFocusColor(s.Frame().Border.FocusColor.Color())
}

// Name returns the component name.
func (*AIChat) Name() string { return aiTitle }

// Start starts the view updater.
func (*AIChat) Start() {}

// Stop terminates the updater.
func (c *AIChat) Stop() {
	c.app.Styles.RemoveListener(c)
}

// Hints returns menu hints.
func (c *AIChat) Hints() model.MenuHints {
	return c.actions.Hints()
}

// ExtraHints returns additional hints.
func (*AIChat) ExtraHints() map[string]string {
	return nil
}

// App returns the app reference.
func (c *AIChat) App() *App {
	return c.app
}

// GetContext returns the context from app.
func (c *AIChat) GetContext() context.Context {
	return context.WithValue(context.Background(), internal.KeyApp, c.app)
}
