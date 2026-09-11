## 1. CLI foundation and discovery

- [x] 1.1 Create the Go module, `ban-code-comments` entry point, option parser, typed scan model, and status mapping; verify `go test ./...` passes and clean/finding/error command fixtures return 0/1/2.
- [x] 1.2 Implement the supported-language registry for extensions and special filenames, language filters, and category selection; verify table-driven registry tests cover every documented language and invalid selections.
- [x] 1.3 Implement recursive repository-aware discovery with `.gitignore`, fixed VCS/dependency/build exclusions, include/exclude globs, unsupported-file skipping, and stderr debug diagnostics; verify temporary-repository integration tests cover precedence and output-stream separation.

## 2. Comment detection and classification

- [x] 2.1 Implement reusable pure-Go lexical state machines for C-style, hash-style, shell, Python, HTML/XML, SQL, CSS, and format-specific comment syntaxes; verify fixtures reject delimiters in strings, raw strings, templates, escaped text, heredocs, and triple-quoted literals.
- [x] 2.2 Connect every supported language to the appropriate scanner family and source-range reporting; verify language fixtures detect ordinary comments with correct path, line, column, and text.
- [x] 2.3 Implement ordinary, documentation, header, and directive classification with practical default categories; verify fixtures cover default suppression and explicit inclusion of headers and directives.

## 3. Reporting and command behavior

- [x] 3.1 Implement the stable JSON result object with findings and summary plus `--format text`; verify golden tests cover empty, finding, and text-mode results.
- [x] 3.2 Complete end-to-end CLI coverage for default current-directory scanning, supplied paths, language/category filters, invalid options, unreadable paths, and debug mode; verify all integration tests pass through the built binary.

## 4. Release and documentation

- [x] 4.1 Add GoReleaser configuration and a least-privilege, tag-triggered GitHub Actions release workflow for the required macOS, Linux, and Windows targets with checksums; verify `goreleaser check` and a local snapshot build succeed.
- [x] 4.2 Update the README with installation, JSON/text output, flag reference, exit-code contract, language/category coverage, and GitHub Actions/Claude hook invocation examples; verify documented commands match CLI integration tests.

## 5. Final verification

- [ ] 5.1 Run the full Go test suite, static checks, formatting validation, OpenSpec strict validation, and GoReleaser validation; verify the tracked diff contains only the approved CLI, release, documentation, and OpenSpec artifacts.
