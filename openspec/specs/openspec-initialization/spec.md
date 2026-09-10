# openspec-initialization Specification

## Purpose
Provides a shared, discoverable OpenSpec workflow so agents and contributors can plan, implement, verify, and archive repository changes consistently.

## Requirements

### Requirement: Shared OpenSpec workflow is available

The repository SHALL provide the OpenSpec project configuration and shared agent skills required to discover and execute the spec-driven workflow.

#### Scenario: Agent discovers the shared workflow

- **WHEN** an agent inspects the repository for OpenSpec guidance
- **THEN** it can find the OpenSpec configuration and shared skills for exploring, proposing, applying, updating, verifying, syncing, and archiving changes

### Requirement: OpenSpec artifacts use the spec-driven schema

The repository SHALL configure OpenSpec to create and manage changes using the `spec-driven` schema.

#### Scenario: New change uses the configured schema

- **WHEN** a contributor creates a new OpenSpec change
- **THEN** the change is resolved against the repository's `spec-driven` configuration
