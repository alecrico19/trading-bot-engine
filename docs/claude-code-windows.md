# Claude Code on Windows — Options

## The Problem

Your project lives in WSL (Linux): `/home/alecr/trading-bot/`

The Claude Windows desktop app is a chat interface — it can't edit files in your WSL filesystem. You need a tool that CAN.

## Option 1: Claude Code CLI in WSL (Recommended)

Same as we discussed — runs inside WSL, has direct file access:

```bash
# In your WSL terminal
npm install -g @anthropic-ai/claude-code
cd ~/trading-bot
claude
```

This is the simplest path. Claude Code CLI runs in your existing WSL environment where everything already works.

## Option 2: VS Code + Claude Extension

If you want a GUI editor on Windows:

1. Install VS Code on Windows
2. Install the "WSL" extension (auto-connects to your WSL files)
3. Search for "Claude" extensions in VS Code marketplace
4. Open `~/trading-bot` folder from VS Code → WSL
5. Claude extension reads the project files through VS Code's WSL integration

## Option 3: Cursor

[Cursor](https://cursor.sh) is an AI-first code editor with Claude support. It can open WSL folders like VS Code.

## Option 4: Continue.dev

Open-source AI assistant for VS Code. Supports Claude via API key. Works with WSL projects.

## What Won't Work

- **Claude Desktop App (Windows)** — chat-only, can't edit files
- **Claude Web (claude.ai)** — can't access your local filesystem

## Recommendation

**Option 1 (CLI in WSL)**. Zero setup — everything already works. You paste the context summary, Claude picks up right where we left off. Same terminal, same commands, same workflow.