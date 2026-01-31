package templates

// LessonTemplate is the template for creating new lessons
const LessonTemplate = `# Title of the lesson

## Symptom

Describe what you observed (error message, unexpected behavior, etc.)

## Root Cause

Explain why this happened

## Fix

What solved the problem

## Prevention

How to avoid this in the future
`

// DecisionTemplate is the template for creating new decisions
const DecisionTemplate = `# Decision Title

## Context

What is the context for this decision? What problem are you trying to solve?

## Decision

What did you decide?

## Rationale

Why did you make this decision?

## Alternatives Considered

What other options did you consider and why were they rejected?

## Consequences

What are the implications of this decision?
`

// PatternTemplate is the template for creating new patterns
const PatternTemplate = `# Pattern Name

## Pattern

Describe the pattern or convention

## Rationale

Why is this pattern used?

## Implementation

How to implement this pattern

## Examples

Show examples of correct usage

## Anti-Patterns

What to avoid
`
