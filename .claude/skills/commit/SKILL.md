---
name: commit
description: Create a git commit following the project's commit message style
allowed-tools: Bash(git *)
---

Create a git commit:

1. Run `git status` and `git diff` to review all changes
2. Run `git log --oneline -5` to check recent commit message style
3. Stage relevant files (avoid staging secrets, .env, or binary files)
4. Write a concise commit message matching the project's existing style (imperative mood, no prefix convention)
5. Commit using a HEREDOC:

```bash
git commit -m "$(cat <<'EOF'
<message>

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

6. Run `git status` to verify success

Rules:
- NEVER amend existing commits unless explicitly asked
- NEVER use --no-verify
- If pre-commit hooks fail, fix the issue and create a NEW commit
- Do not push unless explicitly asked
