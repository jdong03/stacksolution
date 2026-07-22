// Package kuhn implements Kuhn poker: a 3-card, single-betting-round poker
// game small enough to solve exactly. It exists to verify the CFR
// fundamentals (regret matching, average-strategy convergence, info-set
// keying, zero-sum sign conventions) in isolation, decoupled from the
// larger action_tree/ engine - see docs/ROADMAP.md Phase 1.
//
// Rules: each of the two players antes 1 chip and is dealt one private card
// from a 3-card deck {J, Q, K}. Player 1 (index 0) acts first. Each player
// may only Pass (check/fold) or Bet/Call 1 chip; there is no raising.
// Higher card wins at showdown.
package kuhn

// Card ranks the 3-card Kuhn deck. Higher value wins at showdown; ties are
// impossible since each deal draws two distinct cards.
type Card int

const (
	Jack Card = iota
	Queen
	King
)

func (c Card) String() string {
	switch c {
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return "?"
	}
}

// Deck is the fixed 3-card Kuhn deck.
var Deck = []Card{Jack, Queen, King}

// Action is either Pass (check if no bet is outstanding, fold if facing one)
// or Bet (bet if no bet is outstanding, call if facing one).
type Action byte

const (
	Pass Action = 'p'
	Bet  Action = 'b'
)

// Actions lists the two legal actions in play order; index position doubles
// as the action index used by InfoSet's Regret/StrategySum slices.
var Actions = [2]Action{Pass, Bet}

// terminalHistories enumerates every history string at which the hand ends.
var terminalHistories = map[string]bool{
	"pp": true, "bb": true, "bp": true, "pbb": true, "pbp": true,
}

// IsTerminal reports whether history is a completed hand.
func IsTerminal(history string) bool {
	return terminalHistories[history]
}

// activePlayer returns which player (0 or 1) acts next given history.
// Player 0 acts on even action counts, Player 1 on odd - Player 1 (index 0)
// always acts first.
func activePlayer(history string) int {
	return len(history) % 2
}

// TerminalUtility returns the payoff to Player 1 (index 0) at a terminal
// history, given the two dealt cards (cards[0] = Player 1's card, cards[1] =
// Player 2's card). Player 2's utility is always the negation (zero-sum).
func TerminalUtility(history string, cards [2]Card) float64 {
	p1Higher := cards[0] > cards[1]
	switch history {
	case "pp":
		if p1Higher {
			return 1
		}
		return -1
	case "bb", "pbb":
		if p1Higher {
			return 2
		}
		return -2
	case "bp":
		// Player 2 folded to Player 1's bet - Player 1 wins Player 2's ante
		// outright, regardless of cards.
		return 1
	case "pbp":
		// Player 1 folded after checking then facing Player 2's bet -
		// Player 1 loses their ante outright, regardless of cards.
		return -1
	default:
		panic("kuhn: TerminalUtility called on non-terminal history " + history)
	}
}
