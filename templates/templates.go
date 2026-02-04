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

// ConventionTemplate is the template for creating new conventions
const ConventionTemplate = `# Convention Name

## Convention

Describe the standard or convention

## Rationale

Why is this convention used?

## Implementation

How to follow this convention

## Examples

Show examples of correct usage

## Exceptions

When this convention doesn't apply
`
