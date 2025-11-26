# Projection Tables

## Projection tables

Projection functions translate immutable events into read-optimized tables. They assume events

arrive in Lamport order (enforced by rebuild paths) and prefer idempotent writes so replays can

safely rerun after migrations.

Projection Functions

These functions project events from the events table into projection tables

*Source: `tk/internal/database/projections.go:12`*

