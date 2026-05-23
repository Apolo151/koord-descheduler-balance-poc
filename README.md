# koord-descheduler-balance-poc

A Proof-of-Concept Go implementation of an intra-node resource imbalance analyzer and scale-down binpacking simulator for Koordinator.

## Overview

This PoC models nodes, pods, and resources, then evaluates:
- Intra-node imbalance using standard deviation across resource ratios.
- Eviction candidates that most reduce imbalance (delta std dev).
- Scale-down binpack decisions that concentrate remaining pods on fewer nodes.

The project is a standalone simulator with fixtures, deterministic outputs, and unit/integration-style tests.

## Goals

- Detect nodes whose per-resource allocation imbalance exceeds a threshold.
- Rank evictable pods by the reduction in imbalance after eviction.
- Choose scale-down victims that reduce fragmentation and improve node packing.
- Provide fixtures, golden outputs, and documented results.

## Project layout

```
koord-descheduler-balance-poc/
	cmd/
		poc/                 # CLI entrypoint; wiring only
	pkg/
		model/               # Core domain types and helpers
		imbalance/           # Std dev math and delta scoring
		descheduler/         # Imbalance auditor + eviction planning
		binpack/             # Scale-down victim selection
		fixtures/            # YAML/JSON loading + validation
		report/              # Human-readable summaries
	testdata/
		scenarios/           # Fixture inputs
		golden/              # Expected outputs
	README.md
	go.mod
```

## Configuration and contracts

- Resource quantities use float64 (CPU in cores, memory in bytes or MiB).
- GPU values are integer-only (stored as float64, validated at load time).
- Scale-down targets support both workload name and label selector.
- `kube-system` is protected by default in the eviction policy.

## How to run
```
go run ./cmd/poc --scenario testdata/scenarios/basic-imbalance.yaml
```

## Testing

Planned test coverage:
- Unit tests for scoring math and ranking stability.
- Integration-style tests that load fixtures and compare against golden outputs.

## Notes

This PoC is intentionally independent of Koordinator runtime APIs. It models policies like ReservationFirst as config flags and focuses on deterministic evaluation and repeatable results.
