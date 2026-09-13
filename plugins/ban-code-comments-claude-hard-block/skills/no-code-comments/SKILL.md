---
name: no-code-comments
description: Keep supported source and configuration files free of comments while preserving useful documentation in the right place.
---

# No Code Comments

Code comments are prohibited by the supported `ban-code-comments` scanner policy. Write Git history for history, pull-request descriptions for rationale and context, simplified code for complexity, nearby README files for directory-specific explanations, and Markdown for general documentation.

Do not add comments that merely restate readable code. String literals, raw strings, templates, escaped literals, heredocs, triple-quoted literals, Markdown, README files, and unsupported file types remain valid documentation paths when they are appropriate.

The hard-block plugin denies supported `Edit` and `Write` calls that introduce findings. Bash, PowerShell, MCP, NotebookEdit, generators, redirection, and other opaque writes are enforced by the repository scanner and GitHub Action in CI. Review the finding, remove the comment, or move the information to an approved documentation alternative.

Operational hook failures allow the edit with a diagnostic; successfully evaluated findings remain denied.
