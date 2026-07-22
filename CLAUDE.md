# CLAUDE.md

Operating manual for agents working in this repo. Keep this file lean — it loads
every session, so every line costs context. The project plan and current status
live in `docs/ROADMAP.md`; read that before starting phase work.

@docs/ROADMAP.md

## What this is

StackSolution is a poker solver in Go. It computes approximate Nash-equilibrium
strategies for imperfect-information poker games using counterfactual regret
minimization (CFR). Current focus is **correctness on small games (Kuhn, Leduc)**
before any scaling.

## Build / test / verify

- Build: `go build ./...`
- Test: `go test ./...`
- Single test: `go test ./path/to/pkg -run TestName -v`
- Race check (when touching shared solver state): `go test -race ./...`
- Before committing: `go vet ./...` and `gofmt -w .`
- Benchmarks (perf work only): `go test -bench=. -benchmem ./...`

## Prime directive: verify against ground truth, never vibes

This is a correctness-critical numerical algorithm. A plausible-looking CFR
implementation can be subtly wrong and still *look* like it is converging. So:

- **Never mark a solver task complete unless the exploitability eval runs and
  trends toward zero.** "It compiles and the code looks right" is not done.
  "Exploitability on Kuhn is dropping toward ~0" is done.
- When you change anything in the solver, regret, or strategy-averaging code, run
  the exploitability eval and **report the number**, not just that tests pass.
- If exploitability plateaus above the expected range (see ROADMAP reference
  table), assume a bug and investigate before adding features.
- Prefer small, verifiable commits — one conceptual change each — so a regression
  in exploitability can be bisected.

## CFR correctness gotchas (where bugs actually hide)

Check these explicitly; they are the classic ways CFR goes quietly wrong:

- **Two different reach weightings.** Counterfactual value / regret updates weight
  by the *opponent's (and chance's) reach probability* — NOT the acting player's
  own reach. The *average-strategy* accumulation weights by the acting player's
  *own* reach probability. Swapping these two is the single most common CFR bug.
- **The average strategy converges, not the current one.** The equilibrium is the
  average strategy over all iterations. Using/reporting the final current strategy
  is wrong.
- **Info-set keys must not leak hidden information.** Key an info set by (acting
  player's observed cards + public betting history) only. If the opponent's
  private card can influence the key, you are accidentally solving a
  perfect-information game and exploitability will look suspiciously perfect.
- **Zero-sum sign conventions.** Player 2's utility is the negation of player 1's.
  A sign error here produces confident "convergence" to nonsense.
- **Chance nodes.** Card deals are chance nodes and must be weighted by their
  probabilities during traversal.
- **Pick one variant and stay consistent.** Vanilla CFR keeps cumulative regrets
  (which can go negative); CFR+ floors them at zero and averages differently.
  Don't mix conventions inside one solver.

## Architecture (verified against the actual code 2026-07-22)

- `game/` — card primitives only: `Card`/`Deck` (`deck.go`), 7-card hand
  evaluation (`hand_evaluation.go`), and a bare `Player` struct. `game.go` is
  an empty placeholder (just a TODO comment). Betting/action logic, legal
  actions, and payoffs are **not** here — they live in `action_tree/`. There is
  no game-abstraction interface yet; "new games go here behind one common
  interface" is a goal for Phase 1/2, not current state.
- `action_tree/` — this is where almost everything actually lives:
  - Game tree nodes: `PlayerNode`/`ChanceNode`/`LeafNode` (`player_node.go`,
    `chance_node.go`, `leaf_node.go`, `game_state_node.go`), betting history
    and legal-action generation (`history.go`).
  - The solver core: information sets, regret matching, and strategy
    averaging (`information_set.go`), and the CFR training loop itself
    (`trainer.go`).
  - The exploitability oracle: `best_response_utility.go` already implements
    a best-response/`TotalDeviation` calculator and is wired into
    `trainer.Train()`, printing exploitability every iteration. Phase 1's
    "build the eval harness" checkpoint is effectively done — it just hasn't
    been pointed at Kuhn yet.
  - Supporting tooling: `metrics.go` (CSV export for graphing),
    `heuristic_eval.go`/`ev_analysis.go` (compare solver output to hand-coded
    heuristics), `strategy_display.go` (CLI printouts).
- `stacksolution/` — **dead/orphaned code, not the solver core.** Its files
  declare `package game` (not `package stacksolution`) and nothing in the repo
  imports `github.com/jdong03/stacksolution/stacksolution`. It's an early
  duplicate draft of deck/hand-eval/player that was superseded by `game/` but
  never deleted. Do not build on it; consider removing it to avoid confusion.
- `utils/` — shared helpers. Currently just `math_util.go` (`Normalize`), used
  by regret matching in `action_tree/information_set.go`.

Note: the game currently implemented in `game/` + `action_tree/` is **not**
Kuhn or Leduc — it's the earlier simplified project (static board, AA/KK/QQ
ranges only, full flop/turn/river betting with bet-sizing options). No Kuhn or
Leduc implementation exists yet in the repo.

## Conventions

- Idiomatic Go. Keep the game/solver boundary clean so a new game plugs in
  **without touching the solver** — Phase 2 (Leduc) is the test of this.
- Update `docs/ROADMAP.md`'s "Current status" whenever you finish a phase or
  change direction.

Module path: `github.com/jdong03/stacksolution`
