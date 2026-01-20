# Troubleshooting

Common issues and their solutions.

## Installation Issues

### "command not found: bujo"

The binary is not in your PATH.

**Homebrew install:**
```bash
# Ensure Homebrew bin is in PATH
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc
source ~/.zshrc
```

**Go install:**
```bash
# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

**Manual install:**
```bash
# Move to a directory in PATH
sudo mv bujo /usr/local/bin/
```

### Homebrew: "No formulae or casks found"

Tap the repository first:
```bash
brew tap typingincolor/tap
brew install bujo
```

## Database Issues

### "database is locked"

Another process is using the database.

1. Check for running bujo processes:
   ```bash
   ps aux | grep bujo
   ```

2. If using cloud sync (Dropbox, iCloud), ensure the file isn't being synced

3. Kill any stale processes:
   ```bash
   killall bujo
   ```

### "no such table: entries"

The database hasn't been initialized or is corrupted.

```bash
# Remove and recreate (loses data)
rm ~/.bujo/bujo.db
bujo today  # Creates fresh database

# Or restore from backup
bujo backup
bujo import ~/.bujo/backups/bujo_2026-01-15.db --mode replace
```

### Database corruption

If SQLite reports corruption:

1. Try to recover:
   ```bash
   sqlite3 ~/.bujo/bujo.db ".recover" | sqlite3 ~/.bujo/recovered.db
   mv ~/.bujo/recovered.db ~/.bujo/bujo.db
   ```

2. Or restore from backup:
   ```bash
   cp ~/.bujo/backups/bujo_latest.db ~/.bujo/bujo.db
   ```

## TUI Issues

### TUI doesn't render correctly

**Wrong terminal size:**
- Resize your terminal window
- Minimum recommended: 80x24

**Missing Unicode support:**
- Use a terminal with Unicode support (iTerm2, Alacritty, Windows Terminal)
- Ensure locale is set: `export LANG=en_US.UTF-8`

**Colors not showing:**
- Set TERM variable: `export TERM=xterm-256color`
- Check terminal color support: `tput colors` (should be 256)

### Keyboard shortcuts not working

**In tmux/screen:**
- Some key combinations are intercepted
- Use raw mode: `Ctrl+b :set -g mouse off`

**On macOS:**
- Check System Preferences > Keyboard > Shortcuts for conflicts

## AI Features

### "AI features disabled"

Enable AI in configuration:
```bash
export BUJO_AI_ENABLED=true
```

### Local AI not responding

1. Check Ollama is running:
   ```bash
   ollama list
   ```

2. Pull required model:
   ```bash
   ollama pull llama3.2:3b
   ```

3. Verify connection:
   ```bash
   curl http://localhost:11434/api/tags
   ```

### Gemini API errors

**"API key not valid":**
- Verify key is set: `echo $GEMINI_API_KEY`
- Get a new key from [Google AI Studio](https://aistudio.google.com/)

**"Quota exceeded":**
- Wait for quota reset (usually daily)
- Consider local AI as alternative

**Rate limiting:**
- Reduce frequency of summary requests
- Cached summaries avoid repeated API calls

## Command Issues

### "entry not found"

The entry ID may refer to an old version.

```bash
# Check if entry was deleted
bujo deleted

# Try searching for it
bujo search "partial content"
```

### "cannot migrate note"

Only tasks can be migrated. Change the entry type first:
```bash
bujo retype 42 task
bujo migrate 42 --to tomorrow
```

### Date parsing errors

Use explicit formats if natural language fails:
```bash
# Instead of: bujo add --date "next tuesday"
bujo add --date 2026-01-28 ". My task"
```

Supported formats:
- ISO: `2026-01-20`
- Natural: `today`, `tomorrow`, `yesterday`
- Relative: `last monday`, `next friday`

## Performance Issues

### Slow startup

Large databases can slow startup.

1. Archive old versions:
   ```bash
   bujo archive --older-than 2025-01-01 --execute
   ```

2. Vacuum the database:
   ```bash
   sqlite3 ~/.bujo/bujo.db "VACUUM"
   ```

### TUI lag

1. Reduce visible entries by using filters
2. Close other terminal tabs/windows
3. Check for resource-heavy processes

## Desktop App Issues

### App won't start

1. Check for error logs:
   ```bash
   # macOS
   /Applications/Bujo.app/Contents/MacOS/bujoapp 2>&1
   ```

2. Reset preferences:
   ```bash
   rm -rf ~/Library/Application\ Support/bujo/
   ```

### "App is damaged" on macOS

Remove quarantine attribute:
```bash
xattr -d com.apple.quarantine /Applications/Bujo.app
```

Or allow in System Preferences > Security & Privacy.

## Getting Help

If these solutions don't help:

1. Check existing issues: [GitHub Issues](https://github.com/typingincolor/bujo/issues)
2. Open a new issue with:
   - bujo version (`bujo version`)
   - Operating system and version
   - Steps to reproduce
   - Error messages (if any)
