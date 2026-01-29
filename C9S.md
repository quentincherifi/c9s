# C9S - Claude AI Integration for K9s

C9S brings Claude AI assistance directly into K9s, helping you troubleshoot and understand your Kubernetes resources.

## Features

- **Context-Aware AI**: Claude understands your current cluster, namespace, and selected resources
- **Interactive Chat**: Ask questions about your Kubernetes environment
- **Troubleshooting Help**: Get explanations and solutions for common issues

## Configuration

### Set API Key

From any K9s view (pods, deployments, etc.), run:

```
:claude set-key YOUR_ANTHROPIC_API_KEY
```

Or configure in `~/.config/k9s/config.yaml`:

```yaml
k9s:
  ai:
    enabled: true
    apiKey: "sk-ant-api03-..."
    # Or use environment variable
    apiKeyEnv: "ANTHROPIC_API_KEY"
    model: "claude-sonnet-4-20250514"
    maxTokens: 4096
```

### Environment Variable

You can also set the `ANTHROPIC_API_KEY` environment variable:

```bash
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

## Usage

### Open Claude Chat

```
:claude
```

Or with a question:

```
:claude why is my pod crashing?
```

You can also use the alias:

```
:ai why are there so many restarts?
```

### Keyboard Shortcuts (in Claude view)

| Key | Action |
|-----|--------|
| `:` | Enter prompt mode |
| `Enter` | Send message |
| `Ctrl+L` | Clear chat history |
| `Escape` / `q` | Go back |

## Context Information

When you open the Claude view, it automatically captures:

- **Cluster**: Current cluster name
- **Context**: Active Kubernetes context
- **Namespace**: Current namespace
- **View**: The resource type you were viewing
- **Selected**: The resource you had selected (if any)

This context is sent to Claude so it can provide relevant answers.

## Example Questions

- "Why is this pod in CrashLoopBackOff?"
- "How do I increase the memory limit for this deployment?"
- "What's causing the high restart count?"
- "Explain this service's configuration"
- "How do I debug networking issues between these pods?"

## Privacy

Your API key is stored locally in your K9s configuration file. Queries are sent directly to Anthropic's API.
