## Context

The repository currently provides a pure-Go CLI with JSON/text reporting, exit statuses `0`, `1`, and `2`, and GoReleaser archives for Linux and macOS amd64/arm64 plus Windows amd64. It has no Action metadata or JavaScript package. The Action must preserve the CLI as the canonical policy engine and must work without requiring Go in consumer repositories.

## Goals / Non-Goals

**Goals:**

- Provide a root-level Action consumable as `justindfuller/ban-code-comments@<release>` after checkout.
- Download, verify, cache, and execute the matching released CLI on every supported runner target.
- Translate Action inputs to the existing CLI argument contract without shell interpolation.
- Keep release coupling, permissions, documentation, and hosted validation explicit.

**Non-Goals:**

- Rewriting the Go scanner, changing supported-language behavior, or creating a second detector.
- Adding a repository configuration file, annotations, Action-specific finding outputs, or a new report schema.
- Supporting platforms for which GoReleaser does not publish an archive.

## Decisions

### Use a thin JavaScript Action around the Go binary

The root `action.yml` will use a supported Node runtime and a committed bundled JavaScript entrypoint. A JavaScript Action is preferred over a composite shell Action because it can handle platform detection, archive extraction, checksums, caching, argument boundaries, and Windows execution consistently. A Docker Action is rejected because it would limit consumers to Linux runners.

### Couple each Action release to one exact CLI release

The Action source will contain one embedded CLI version constant and will not expose a runtime version override. Release validation will require that the constant matches the semver release tag. Exact tags therefore select a reproducible Action/CLI pair, while an optional moving major tag can point to the latest compatible exact release for convenient adoption.

### Use explicit release asset names and checksum verification

GoReleaser will use an explicit archive name template based on project, version, operating system, and architecture. The downloader will construct the corresponding release URL, fetch `checksums.txt`, verify the archive with SHA-256, and only then extract it. The extracted directory will be searched for the expected executable so the current wrapped archive layout remains supported.

### Cache the verified executable

The Action will use the GitHub Actions tool-cache mechanism keyed by CLI version and runner target. A cache hit will still identify the expected executable, while a newly downloaded archive must pass checksum verification before being cached or executed.

### Translate inputs into an argument array

The Action will read `paths`, `languages`, `categories`, `include`, `exclude`, `format`, and `debug`, apply the documented defaults, and construct an argument array for the CLI. Paths are newline-separated; language/category values preserve the CLI's comma-separated syntax; include and exclude entries may be newline-separated. The executable will be spawned without a shell and with the consumer workspace as its working directory.

### Preserve process streams and statuses

The child process will inherit or stream standard output and standard error so JSON/text reports and debug diagnostics remain the same as direct CLI execution. Its exit code will be returned unchanged, allowing clean scans, policy findings, and operational failures to remain distinct even though both nonzero statuses fail a workflow step.

### Treat the Action as read-only in consumer repositories

The Action will not use `GITHUB_TOKEN`, call repository APIs, modify files, or perform a checkout. The README will require callers to check out their repository first and will show `contents: read` permissions. The Action's own workflow dependencies will remain pinned to full commit SHAs, and the release process will publish the bundled entrypoint required by tagged consumers.

## Risks / Trade-offs

- [A release asset can be missing or renamed] -> Make GoReleaser naming explicit and add release-artifact tests for every supported target.
- [A checksum or archive can be unavailable due to transient network failure] -> Produce actionable errors, use the tool cache after successful verification, and validate the real Action in hosted workflows.
- [A moving major tag reduces reproducibility] -> Make exact semver references canonical in documentation and recommend full commit-SHA pinning for strict supply-chain policies.
- [Windows or self-hosted runner behavior can differ from hosted Linux] -> Test supported operating systems in hosted matrix coverage and keep invocation shell-free.
- [The checked-in bundled entrypoint can drift from source] -> Build it in CI and fail when the generated distribution differs from the committed files.

## Migration Plan

1. Add the Action metadata, downloader, input adapter, bundled entrypoint, tests, and documentation without changing the Go CLI contract.
2. Make release asset naming and Action-version validation part of CI/release checks.
3. Publish the first exact semver Action/CLI release, then update the documented major alias if used.
4. Roll back a faulty Action by directing consumers to the previous exact release or commit SHA; no consumer repository migration is required.
