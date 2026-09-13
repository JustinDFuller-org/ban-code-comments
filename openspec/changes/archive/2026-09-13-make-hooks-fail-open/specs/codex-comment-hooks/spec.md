## ADDED Requirements

### Requirement: Codex hook operational failures fail open
The Codex plugins SHALL allow the tool call when hook input cannot be decoded or a proposed edit cannot be evaluated, while returning a visible non-blocking diagnostic that identifies the operational error.

#### Scenario: Normal stdin event is decoded
- **WHEN** a Codex plugin receives a valid JSON `PreToolUse` event on stdin
- **THEN** it evaluates the event rather than treating the stdin stream as the event value

#### Scenario: Unevaluable event does not block
- **WHEN** a Codex plugin cannot decode or evaluate a hook event or proposed edit
- **THEN** it emits a diagnostic without a hard-block decision and the underlying tool call proceeds
