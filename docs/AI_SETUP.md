# AI Summary Setup Guide

This guide explains how to set up Google Gemini API for AI-powered summaries in bujo.

## Prerequisites

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

## Troubleshooting

### "GEMINI_API_KEY environment variable is required"

The API key is not set. Check:
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

## Security Notes

- Never commit `.env` files to version control
- Add `.env` to your `.gitignore`
- Use restrictive file permissions: `chmod 600 ~/.bujo/.env`
- Consider using Google Cloud's API key restrictions to limit usage

## API Costs

As of 2025, Gemini API has a generous free tier:
- Gemini 2.0 Flash: Free for most personal use
- Check [Google AI pricing](https://ai.google.dev/pricing) for current rates

## Model Used

bujo uses `gemini-2.0-flash` by default, which provides:
- Fast response times
- Good quality summaries
- Cost-effective for frequent use
