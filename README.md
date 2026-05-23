# koord-descheduler-balance-poc

A Proof-of-Concept Go implementation of an intra-node resource imbalance analyzer and scale-down binpacking simulator for Koordinator.

## Overview

This PoC models nodes, pods, and resources, then evaluates:
- Intra-node imbalance using standard deviation across resource ratios.
- Eviction candidates that most reduce imbalance (delta std dev).
- Scale-down binpack decisions that concentrate remaining pods on fewer nodes.

The project is a standalone simulator with fixtures, deterministic outputs, and unit/integration-style tests.

## Architecture

- Descheduler flow: score each node's imbalance, flag nodes above threshold, rank evictable pods by delta std dev, and produce eviction plans.
- Binpack flow: select scale-down victims that reduce the number of nodes hosting target pods, using deterministic tie-breakers.
- Shared utilities: resource ratio math, standard deviation, and fixture-driven inputs.

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

Write outputs:
```
go run ./cmd/poc --scenario testdata/scenarios/basic-imbalance.yaml --out-json out.json --out-summary out.txt
```

## Fixture schema (minimal example)

```
metadata:
	name: example
cluster:
	pods:
		default/pod-a:
			owner: app
			resources:
				cpu: 2
				memory: 4
			labels:
				app: demo
			evictable: true
	nodes:
		node-a:
			allocatable:
				cpu: 10
				memory: 10
			pods:
				- default/pod-a
imbalanceConfig:
	resources: [cpu, memory]
	threshold: 0.1
evictionPolicy:
	reservationFirst: true
	denyNamespaces: [kube-system]
binpackConfig:
	targetWorkload: app
	victims: 1
events:
	scaleDown:
		targetWorkload: app
		victims: 1
```

Example fixtures:
- [testdata/scenarios/basic-imbalance.yaml](testdata/scenarios/basic-imbalance.yaml)
- [testdata/scenarios/scale-down-binpack.yaml](testdata/scenarios/scale-down-binpack.yaml)

## Metrics

Per-node imbalance uses standard deviation across resource ratios. For CPU and memory ratios $f_{cpu}$ and $f_{mem}$:

$$
\mu = \frac{f_{cpu} + f_{mem}}{2}
$$

$$
\sigma = \sqrt{\frac{(f_{cpu} - \mu)^2 + (f_{mem} - \mu)^2}{2}}
$$

Delta scoring ranks pods by reduction in $\sigma$ after hypothetical removal.

## Testing

Run all tests:
```
go test ./...
```

Coverage includes:
- Unit tests for scoring math, eviction filtering, and binpack selection.
- Integration tests that load fixtures and assert victim selection.
- Golden summary tests against [testdata/golden](testdata/golden).

## Benchmark summary

This PoC reports before/after fragmentation metrics from the fixture outputs:

| Metric | Before | After | Improvement |
| --- | --- | --- | --- |
| Nodes above imbalance threshold | 1 | 0 | 100% |
| Nodes hosting target pods (scale-down) | 3 | 2 | 33% |

Interpretation:
- Descheduler evictions reduce per-node imbalance.
- Binpack selection concentrates target pods onto fewer nodes.

## Notes

This PoC is intentionally independent of Koordinator runtime APIs. It models policies like ReservationFirst as config flags and focuses on deterministic evaluation and repeatable results.
