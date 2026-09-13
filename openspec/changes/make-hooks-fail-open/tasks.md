## 1. Input and error handling

- [ ] 1.1 Normalize Codex hook input so stdin streams are read before JSON decoding and already-decoded programmatic events remain supported; verify with direct runner tests.
- [ ] 1.2 Update Codex launcher-level mode, input, and evaluation failures to emit non-blocking diagnostics; verify valid stdin events reach hard-mode finding enforcement.
- [ ] 1.3 Update Claude hook operational-error and launcher failure paths to emit non-blocking native diagnostics; verify malformed and unevaluable events proceed.
- [ ] 1.4 Preserve hard-mode denial and warn-mode guidance for successfully evaluated findings; verify existing finding tests remain green.

## 2. Regression coverage

- [ ] 2.1 Add Codex tests for valid stdin, malformed JSON, malformed proposals, invalid modes, and diagnostic response shape; verify no operational case contains a block decision.
- [ ] 2.2 Add Claude tests for malformed JSON, malformed proposals, invalid modes, and diagnostic response shape; verify no operational case contains a deny decision.
- [ ] 2.3 Add source and generated-bundle launcher smoke coverage for clean, finding, malformed, and unevaluable events; verify all four plugin variants behave consistently.

## 3. Documentation and distribution

- [ ] 3.1 Update README and plugin guidance to explain that operational hook failures fail open with diagnostics while repository scanning and CI remain authoritative; verify documented behavior matches the specs.
- [ ] 3.2 Regenerate the Codex and Claude plugin launcher bundles from source; verify generated artifacts contain the corrected behavior and no unintended files change.

## 4. Validation

- [ ] 4.1 Run the complete Node test suite and plugin validators; verify all tests and validators pass.
- [ ] 4.2 Run strict OpenSpec validation for the completed change; verify every required artifact is valid and complete.
- [ ] 4.3 Run installed-path or isolated plugin smoke tests for both hard and warn variants; verify actual bundled launchers allow operational failures and enforce genuine findings.
