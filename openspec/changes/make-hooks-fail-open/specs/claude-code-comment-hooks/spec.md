## ADDED Requirements

### Requirement: Claude hook operational failures fail open
The Claude Code plugins SHALL allow the tool call when hook input cannot be decoded or a proposed edit cannot be evaluated, while returning a visible non-blocking diagnostic that identifies the operational error.

#### Scenario: Unevaluable event does not block
- **WHEN** a Claude plugin cannot decode or evaluate a hook event or proposed edit
- **THEN** it emits a diagnostic without a deny decision and the underlying tool call proceeds

#### Scenario: Findings remain enforceable
- **WHEN** a supported edit is successfully evaluated and introduces a policy finding
- **THEN** hard mode denies it and warn mode retains its existing warning behavior
