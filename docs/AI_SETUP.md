# AI Setup Guide

This guide explains how to set up AI-powered summaries in bujo. You can choose between:

- **Local AI** (Recommended): Run AI models locally for complete privacy and offline use
- **Google Gemini API**: Cloud-based AI with fast response times

## Enabling AI Features

**AI features are disabled by default.** To enable AI functionality, you must set:

```bash
export BUJO_AI_ENABLED=true
```

Add this to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) or to `~/.bujo/.env`:

```bash
echo "BUJO_AI_ENABLED=true" >> ~/.bujo/.env
```

Once enabled, configure either Local AI or Gemini as described below.

## Quick Start: Local AI (Recommended)

Local AI runs entirely on your machine - no data leaves your computer, works offline, and has no API costs.

### 1. Install Ollama

```bash
brew install ollama
```

### 2. Start Ollama Service

```bash
# Run without auto-restart on reboot (recommended)
brew services run ollama

# Or run with auto-restart on reboot
brew services start ollama

# Check status
brew services info ollama

# Stop the service
brew services stop ollama
```

### 3. Download a Model

```bash
# List available models
ollama list

# Download the recommended model (2.0 GB) - good balance of quality and speed
ollama pull llama3.2:3b

# Or try the smaller model if low on disk space (1.3 GB) - lower quality
ollama pull llama3.2:1b
```

### 4. Use AI Features

```bash
# Generate summaries
bujo summary daily
bujo summary weekly

# The local model is used automatically
```

That's it! Your AI runs locally with complete privacy.

## Alternative: Google Gemini API

If you prefer cloud-based AI, you can use Google's Gemini API instead.

### Prerequisites

- A Google account
- bujo installed and working

## Step 1: Get a Gemini API Key

### Option A: Google AI Studio (Recommended for personal use)

1. Go to [Google AI Studio](https://aistudio.google.com/)
2. Sign in with your Google account
3. Click **Get API Key** in the left sidebar
4. Click **Create API Key**
5. Select or create a Google Cloud project
6. Copy the generated API key

### Option B: Google Cloud Console (For production/enterprise)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Generative Language API**:
   - Go to APIs & Services > Library
   - Search for "Generative Language API"
   - Click Enable
4. Create credentials:
   - Go to APIs & Services > Credentials
   - Click **Create Credentials** > **API Key**
   - Copy the generated API key
5. (Optional) Restrict the API key to only the Generative Language API

## Step 2: Configure bujo

### Option A: Using .env file (Recommended)

Create a `.env` file in `~/.bujo/`:

```bash
mkdir -p ~/.bujo
echo "GEMINI_API_KEY=your-api-key-here" > ~/.bujo/.env
chmod 600 ~/.bujo/.env  # Restrict permissions
```

Or create `.env` in your current working directory (useful for development):

```bash
echo "GEMINI_API_KEY=your-api-key-here" > .env
```

**Note:** `.env` in current directory takes precedence over `~/.bujo/.env`.

### Option B: Environment variable

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export GEMINI_API_KEY="your-api-key-here"
```

Then reload:

```bash
source ~/.zshrc  # or ~/.bashrc
```

### Option C: Inline (for testing)

```bash
GEMINI_API_KEY="your-api-key-here" bujo summary
```

## Step 3: Verify Setup

Test that the API key is working:

```bash
bujo summary
```

If configured correctly, you'll see an AI-generated summary of your journal entries.

## AI Provider Configuration

### Choosing a Provider

bujo automatically selects the AI provider based on your configuration:

1. **If `BUJO_AI_PROVIDER` is set**: Uses the specified provider (`local` or `gemini`)
2. **If `GEMINI_API_KEY` is set**: Falls back to Gemini
3. **Otherwise**: Tries to use local AI (requires downloaded model)

### Explicit Provider Selection

Force a specific provider:

```bash
# Always use local AI
export BUJO_AI_PROVIDER=local

# Always use Gemini
export BUJO_AI_PROVIDER=gemini
```

### Local Model Configuration

```bash
# Specify which model to use (default: llama3.2:3b)
export BUJO_MODEL=llama3.2:3b
```

Ollama manages model storage automatically in `~/.ollama/models/`.

### Example Configurations

**Local AI only:**
```bash
export BUJO_AI_PROVIDER=local
export BUJO_MODEL=llama3.2:3b
```

**Gemini only:**
```bash
export BUJO_AI_PROVIDER=gemini
export GEMINI_API_KEY=your-api-key-here
```

**Auto-select (Gemini with local fallback):**
```bash
export GEMINI_API_KEY=your-api-key-here
# Automatically uses Gemini when online, local when GEMINI_API_KEY is unset
```

## Troubleshooting

### "AI features are disabled"

AI features are disabled by default. To enable them:

```bash
export BUJO_AI_ENABLED=true
```

Or add to `~/.bujo/.env`:

```bash
echo "BUJO_AI_ENABLED=true" >> ~/.bujo/.env
```

### "model not downloaded"

You're trying to use local AI but haven't downloaded a model yet:

```bash
# See available models
ollama list

# Download one
ollama pull tinyllama
```

### "pull model manifest: 500: internal error"

Ollama's registry servers are temporarily unavailable. This is a server-side issue. Wait a few minutes and retry:

```bash
ollama pull llama3.2:3b
```

### "failed to create Ollama client: ... (is Ollama running?)"

Ollama service isn't running. Start it with:

```bash
brew services run ollama
```

### "GEMINI_API_KEY environment variable is required"

You're trying to use Gemini but the API key isn't set. Check:
- `.env` file exists and contains `GEMINI_API_KEY=...`
- No typos in the variable name
- File permissions allow reading

### "failed to create Gemini client"

- Verify your API key is correct (no extra spaces)
- Check your internet connection
- Ensure the API is enabled in Google Cloud Console

### "no response generated"

- The API call succeeded but returned empty
- Try again - this can be a transient issue
- Check if you have API quota remaining

## Security and Privacy

### Local AI

- **Complete privacy**: Your journal data never leaves your computer
- **Offline use**: Works without internet connection
- **No API keys**: No credentials to manage or leak
- **No data retention**: No third-party stores your data

### Gemini API

- **Cloud processing**: Journal entries are sent to Google's servers
- **API key security**: Never commit `.env` files to version control
- **File permissions**: Use `chmod 600 ~/.bujo/.env`
- **API restrictions**: Consider using Google Cloud's API key restrictions

## Costs

### Local AI

- **Free**: No API costs
- **Disk space**: Models require 0.6-4 GB storage
- **One-time download**: No ongoing costs
- **Hardware**: Runs on any modern computer

### Gemini API

As of 2025, Gemini API has a generous free tier:
- Gemini 2.0 Flash: Free for most personal use
- Check [Google AI pricing](https://ai.google.dev/pricing) for current rates

## Model Information

### Local Models

bujo supports several curated models optimized for journal summaries:

| Model | Size | Quality | Speed | Memory |
|-------|------|---------|-------|--------|
| tinyllama | 637 MB | Poor | Fast | 1 GB RAM |
| llama3.2:1b | 1.3 GB | Fair | Fast | 2 GB RAM |
| llama3.2:3b | 2.0 GB | Good (Recommended) | Medium | 4 GB RAM |
| phi-3-mini | 2.3 GB | Good | Fast | 3 GB RAM |
| mistral:7b | 4.1 GB | Excellent | Slow | 8 GB RAM |

### Gemini Model

bujo uses `gemini-2.0-flash` by default, which provides:
- Fast response times (~1-2 seconds)
- High quality summaries
- Cost-effective for frequent use

## Customizing Prompts

bujo uses customizable prompt templates for AI-generated summaries. On first run, default templates are automatically created in `~/.bujo/prompts/`:

```bash
~/.bujo/prompts/
├── summary-daily.txt
├── summary-weekly.txt
├── summary-quarterly.txt
├── summary-annual.txt
└── ask.txt
```

### Editing Prompts

You can customize any template by editing the files:

```bash
# Edit daily summary prompt
vim ~/.bujo/prompts/summary-daily.txt

# Or use your preferred editor
code ~/.bujo/prompts/summary-weekly.txt
```

### Template Variables

Prompts use Go template syntax with these variables:

- `{{.Entries}}` - Your journal entries for the period
- `{{.Horizon}}` - Summary type (daily/weekly/quarterly/annual)
- `{{.StartDate}}` - Period start date
- `{{.EndDate}}` - Period end date
- `{{.Question}}` - User's question (for ask command, future feature)

### Example Custom Prompt

```
You are analyzing my {{.Horizon}} journal.

Entries:
{{range .Entries}}
- {{.Content}}
{{end}}

Focus on what I accomplished and what needs attention next.
```

### Resetting to Defaults

To restore default prompts, simply delete the files and they'll be recreated on next run:

```bash
rm ~/.bujo/prompts/summary-daily.txt
bujo summary  # Recreates default daily prompt
```
