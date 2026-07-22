# StackSolution Roadmap

The plan and current status for the solver. **Agents: read this before starting
phase work, and update the "Current status" section as you go.** Working
conventions and the verification discipline live in the root `CLAUDE.md`.

## Goal

Build a correct, eval-verified CFR poker solver by climbing a ladder of games,
each with a known correctness checkpoint: **Kuhn → Leduc → a simplified,
abstracted hold'em spot.**

## Non-goals (for now)

- **Not a GTO Wizard clone.** No precomputed solution libraries across millions of
  spots, no trainer UI, no commercial-scale postflop coverage. Those are
  multi-person, multi-year efforts and not where the learning is.
- The realistic target for this project is a correct solver that handles a
  simplified-but-real spot with basic abstraction.

## Guiding principle

Every rung of the ladder has an objective checkpoint — a known game value or an
exploitability that must trend toward zero. **Never climb to the next rung until
the current rung's eval passes.** The exploitability / best-response calculator is
the oracle that makes this discipline possible, so build it early and keep it
unchanged as games get bigger.

## Phases

### Phase 0 — Verification loop (do this first)

Get `go build ./...` and `go test ./...` green. Stand up the eval-harness scaffold
(even before it is correct) so every later change can be measured.
**Checkpoint:** the whole build + test + eval loop runs locally in one command.

### Phase 1 — Kuhn + vanilla CFR + exploitability

Implement Kuhn poker and vanilla CFR, then the exploitability (best-response)
calculator.
**Checkpoints:**

- The game tree produces exactly **12 information sets** for Kuhn.
- Exploitability **trends toward zero** as iterations increase (see reference
  table below).
- Game value to player 1 approaches **−1/18 ≈ −0.0556**.
- Note: Kuhn has a **one-parameter family** of equilibria for player 1. Do **not**
  test by comparing the strategy to a single reference strategy — test
  *exploitability*. Two different-looking strategies both at ~0 exploitability are
  both correct.

### Phase 2 — Leduc

Extend to Leduc poker (the standard research benchmark: larger, two betting
rounds, one public card).
**Checkpoints:**

- Reuses the **same solver and the same exploitability oracle unchanged** — this
  is the real test that the game/solver boundary is clean.
- Exploitability trends down into the low tenths and keeps dropping with more
  iterations.

### Phase 3 — Faster CFR variants

Add CFR+ and/or discounted / linear CFR, and a Monte Carlo variant (external
sampling), each measured against the **same** exploitability evals.
**Checkpoint:** equal-or-better final exploitability reached in materially fewer
iterations or less wall-clock time, with evals still passing.

### Phase 4 — Abstraction + scaling (you stay the architect)

Introduce card abstraction (bucketing), action / bet-size discretization, and
board isomorphism to attack a simplified hold'em spot. These are domain-judgment
calls — **drive them yourself**; use agents to implement optimizations you
specify, not to choose scope.
**Checkpoint:** a defined simplified spot solves to acceptable exploitability in
acceptable time and memory.

### Phase 5 (stretch) — Turn the solver into a tool

Wrap the solver as an MCP server and/or a "coach" agent that queries it and
explains lines. This is where the solver stops being just a solver and becomes an
*environment* for practicing agent/harness skills.

## Ground-truth reference (calibrate evals against these)

Rough exploitability magnitudes for sanity-checking. Direction and order of
magnitude matter more than exact figures — definitions and iteration counts vary:

- **Kuhn:** a well-tuned CFR run reaches ~0.05 and below; a classical / under-tuned
  run lands around ~0.15. Plateauing near ~0.5 means a bug.
- **Leduc:** early / classical runs sit around ~0.37; tuned variants reach ~0.24
  and lower with more iterations.
- **Kuhn game value to P1:** −1/18 ≈ −0.0556.

## Current status

- [x] Phase 0 — verification loop (`go build/test ./...` green; whole suite
  passes)
- [x] Phase 1 — Kuhn + CFR + exploitability calculator
- [ ] Phase 2 — Leduc
- [ ] Phase 3 — faster variants (CFR+, discounted, MCCFR)
- [ ] Phase 4 — abstraction + simplified hold'em spot
- [ ] Phase 5 — tool wrapper / coach agent (stretch)

_Update this checklist and add short notes (dates, exploitability numbers hit,
decisions made) as phases complete._

**2026-07-22 — Phase 1 done (standalone `kuhn/` package).** Built Kuhn poker,
vanilla CFR, and an independent best-response exploitability oracle, fully
decoupled from `action_tree/` (Option A). All three checkpoints hit and locked
in by `kuhn/kuhn_test.go`: exactly **12 information sets**; **exploitability
0.0011** after 200k iterations (target < 0.01); game value to P1 **-0.0538 vs
-1/18 ≈ -0.0556**. Learned strategies match Kuhn's known one-parameter
equilibrium family. Note for later: the first cut of the exploitability oracle
had the classic imperfect-information bug — it maximized per full deal, letting
the "best responder" see the opponent's card — which read ~0.28 exploitability
against an already-correct trainer. Fixed with belief-propagation backward
induction that commits to one action per info set. Lesson: an eval bug looks
exactly like a solver bug; the giveaway was that the *game value* had already
converged to -1/18 while exploitability hadn't.
