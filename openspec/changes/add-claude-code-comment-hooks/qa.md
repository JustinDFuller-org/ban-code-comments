# Claude Code QA

Run date: 2026-09-13

Installed CLI:

- Claude Code `2.1.216`
- `claude plugin validate plugins/ban-code-comments-claude-hard-block`: passed
- `claude plugin validate plugins/ban-code-comments-claude-warn`: passed

Bundled launcher smoke:

- Hard-block `Write` with a new Go comment: passed, nested `PreToolUse` permission decision was `deny`.
- Warn `Write` with a new Go comment: passed, edit remained allowed and nested `additionalContext` was returned.
- Hard-block `Bash` unsupported path: passed, launcher returned `{}`.
- Warn `Bash` unsupported path: passed, launcher returned `{}`.
- Authenticated `claude -p --plugin-dir` smoke: unrun; `CLAUDE_SMOKE_LIVE=1` is required to opt into live model access.

The validation and launcher results above are installed-CLI and deterministic protocol evidence. They do not claim authenticated model-session coverage.
