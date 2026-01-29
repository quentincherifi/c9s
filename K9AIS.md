# K9AIS - K9s with AI Integration

K9AIS brings AI-powered assistance directly into K9s, supporting multiple AI providers to help you troubleshoot and understand your Kubernetes resources.

## Supported Providers

| Provider | Models | API Key Required |
|----------|--------|------------------|
| **Claude** (Anthropic) | claude-sonnet-4, claude-opus-4, etc. | Yes |
| **OpenAI** | gpt-4o, gpt-4-turbo, gpt-3.5-turbo | Yes |
| **Ollama** | llama3.2, mistral, codellama, etc. | No (local) |

## Quick Start

```bash
# Set your API key (Claude is default)
:ai set-key sk-ant-api03-YOUR_KEY

# Or switch to OpenAI
:ai provider openai
:ai set-key sk-YOUR_OPENAI_KEY

# Or use local Ollama (no key needed)
:ai provider ollama

# Open AI chat
:ai

# Ask a question directly
:ai why is my pod crashing?
```

## Configuration

Add to your `~/.config/k9s/config.yaml`:

### Claude (Default)
```yaml
k9s:
  ai:
    enabled: true
    provider: claude
    apiKey: "sk-ant-api03-..."
    # Or use environment variable
    apiKeyEnv: "ANTHROPIC_API_KEY"
    model: "claude-sonnet-4-20250514"
    maxTokens: 4096
```

### OpenAI
```yaml
k9s:
  ai:
    enabled: true
    provider: openai
    apiKey: "sk-..."
    # Or use environment variable
    apiKeyEnv: "OPENAI_API_KEY"
    model: "gpt-4o"
    maxTokens: 4096
    # Optional: Use Azure OpenAI or other compatible endpoint
    baseURL: "https://your-endpoint.openai.azure.com/v1/chat/completions"
```

### Ollama (Local)
```yaml
k9s:
  ai:
    enabled: true
    provider: ollama
    model: "llama3.2"
    # Optional: Custom Ollama URL
    baseURL: "http://localhost:11434"
```

## Environment Variables

| Variable | Provider | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | Claude | Anthropic API key |
| `OPENAI_API_KEY` | OpenAI | OpenAI API key |

## Commands

| Command | Description |
|---------|-------------|
| `:ai` | Open AI chat view |
| `:ai <question>` | Ask a question directly |
| `:ai set-key <key>` | Set API key for current provider |
| `:ai provider <name>` | Switch provider (claude/openai/ollama) |
| `:chat` | Alias for `:ai` |
| `:ask <question>` | Alias for `:ai <question>` |

## Keyboard Shortcuts (in AI view)

| Key | Action |
|-----|--------|
| `:` | Enter prompt mode |
| `Enter` | Send message |
| `Ctrl+L` | Clear chat history |
| `Escape` / `q` | Go back |

## Context Information

The AI assistant automatically captures:

- **Cluster**: Current cluster name
- **Context**: Active Kubernetes context
- **Namespace**: Current namespace
- **View**: The resource type you were viewing
- **Selected**: The resource you had selected (if any)

This context is included in the AI prompt for relevant answers.

## Example Questions

- "Why is this pod in CrashLoopBackOff?"
- "How do I increase the memory limit for this deployment?"
- "What's causing the high restart count?"
- "Explain this service's configuration"
- "How do I debug networking issues between these pods?"
- "What are the best practices for this HPA configuration?"

## Using Ollama (Local AI)

1. Install Ollama: https://ollama.ai
2. Pull a model: `ollama pull llama3.2`
3. Configure K9s:
   ```yaml
   k9s:
     ai:
       enabled: true
       provider: ollama
       model: llama3.2
   ```
4. Use `:ai` in K9s

## Privacy

- API keys are stored locally in your K9s configuration file
- Queries are sent directly to the selected AI provider
- No data collection or telemetry by K9AIS
- Use Ollama for fully local, private AI assistance
