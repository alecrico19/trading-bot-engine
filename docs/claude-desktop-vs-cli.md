# Claude Desktop App vs Claude Code — Clarification

## Different Products

| Feature | Claude Desktop App | Claude Code CLI |
|---------|-------------------|-----------------|
| Chat with AI | ✅ Yes | ✅ Yes |
| Upload files for context | ✅ Projects | ❌ (reads local) |
| Generate code snippets | ✅ Artifacts | ✅ Inline edits |
| Edit your project files | ❌ No | ✅ Yes |
| Run terminal commands | ❌ No | ✅ Yes |
| Access WSL filesystem | ❌ No | ✅ Yes |
| Git integration | ❌ No | ✅ Yes |
| Price | Included in Pro | Included in Pro |

## The Desktop App "Projects" Feature

You CAN upload files to Claude Projects in the desktop app. Claude reads them for context and can answer questions about your code. But it CANNOT:

- Edit those files on your computer
- Create new files in your project
- Run `go build` or any terminal command
- Access your WSL filesystem directly

It's a **read-only reference tool**, not a coding agent.

## What You Need

For actual code editing in your WSL project, you need the **Claude Code CLI**:

```bash
npm install -g @anthropic-ai/claude-code
cd ~/trading-bot
claude
```

This IS included in your Claude Pro subscription — it's just accessed through the terminal, not the desktop app.

## Hybrid Approach (If You Prefer GUI)

1. Use Claude Desktop App for: reading docs, asking questions, planning features
2. Use Claude Code CLI for: actually editing files, running builds, making commits

Or just use the CLI for everything — it's faster and has full access to your project.
