package game

import "sort"

type HandCategory int

const (
	HighCard HandCategory = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
)

// HandRank encodes a comparable ranking for a poker hand.
// Category defines the class (e.g., Straight), and Primary holds tiebreakers in order.
// Always fill Primary with up to 5 ranks, descending significance; unused entries should be 0.
type HandRank struct {
	Category HandCategory
	Primary  [5]int
}

func lessHandRank(a, b HandRank) bool {
	if a.Category != b.Category {
		return a.Category < b.Category
	}
	for i := 0; i < len(a.Primary); i++ {
		if a.Primary[i] != b.Primary[i] {
			return a.Primary[i] < b.Primary[i]
		}
	}
	return false
}

func maxHandRank(a, b HandRank) HandRank {
	if lessHandRank(a, b) {
		return b
	}
	return a
}

func evaluateHand(flop []Card, playerHand []Card, turn Card, river Card) string {
	r := Evaluate7(flop, playerHand, turn, river)
	switch r.Category {
 	case StraightFlush:
 		// Royal is a special case of straight flush with high card Ace (14) and includes Ten
 		if r.Primary[0] == 14 {
 			return "Royal Flush"
 		}
 		return "Straight Flush"
 	case FourOfAKind:
 		return "Four of a Kind"
 	case FullHouse:
 		return "Full House"
 	case Flush:
 		return "Flush"
 	case Straight:
 		return "Straight"
 	case ThreeOfAKind:
 		return "Three of a Kind"
 	case TwoPair:
 		return "Two Pair"
 	case OnePair:
 		return "One Pair"
 	default:
 		return "High Card"
 	}
}

// Evaluate7 computes the best 5-card hand from 7 cards (player 2 + board 5).
func Evaluate7(flop []Card, playerHand []Card, turn Card, river Card) HandRank {
    cards := append(append([]Card{}, playerHand...), append(flop, turn, river)...)
    // Straight Flush
    if ok, high := straightFlushHigh(cards); ok {
        return HandRank{Category: StraightFlush, Primary: [5]int{high}}
    }
    // Four of a Kind
    if ok, quadRank, kicker := fourOfAKind(cards); ok {
        return HandRank{Category: FourOfAKind, Primary: [5]int{quadRank, kicker}}
    }
    // Full House
    if ok, tripRank, pairRank := fullHouse(cards); ok {
        return HandRank{Category: FullHouse, Primary: [5]int{tripRank, pairRank}}
    }
    // Flush
    if ok, top5 := flushTop5(cards); ok {
        return HandRank{Category: Flush, Primary: toFixed5(top5)}
    }
    // Straight
    if ok, high := straightHigh(cards); ok {
        return HandRank{Category: Straight, Primary: [5]int{high}}
    }
    // Trips
    if ok, tripRank, kickers := tripsWithKickers(cards); ok {
        p := [5]int{tripRank}
        copy(p[1:], kickers)
        return HandRank{Category: ThreeOfAKind, Primary: p}
    }
    // Two Pair
    if ok, highPair, lowPair, kicker := twoPairWithKicker(cards); ok {
        return HandRank{Category: TwoPair, Primary: [5]int{highPair, lowPair, kicker}}
    }
    // One Pair
    if ok, pairRank, kickers := onePairWithKickers(cards); ok {
        p := [5]int{pairRank}
        copy(p[1:], kickers)
        return HandRank{Category: OnePair, Primary: p}
    }
    // High Card
    highs := topRanksExcluding(cards, map[int]int{})
    return HandRank{Category: HighCard, Primary: toFixed5(highs[:min(5, len(highs))])}
}

func toFixed5(src []int) [5]int {
    var out [5]int
    for i := 0; i < len(src) && i < 5; i++ {
        out[i] = src[i]
    }
    return out
}

func rankCounts(cards []Card) map[int]int {
    counts := map[int]int{}
    for _, c := range cards {
        counts[c.Rank]++
    }
    return counts
}

func suitsMap(cards []Card) map[string][]Card {
    m := map[string][]Card{}
    for _, c := range cards {
        m[c.Suit] = append(m[c.Suit], c)
    }
    return m
}

func sortedUniqueRanksDesc(cards []Card, aceLow bool) []int {
    present := map[int]bool{}
    for _, c := range cards {
        present[c.Rank] = true
        if aceLow && c.Rank == 14 {
            present[1] = true
        }
    }
    ranks := make([]int, 0, len(present))
    for r := range present {
        ranks = append(ranks, r)
    }
    sort.Slice(ranks, func(i, j int) bool { return ranks[i] > ranks[j] })
    return ranks
}

func straightHigh(cards []Card) (bool, int) {
    ranks := sortedUniqueRanksDesc(cards, true)
    if len(ranks) < 5 {
        return false, 0
    }
    
    run := 1
    for i := 0; i < len(ranks)-1; i++ {
        if ranks[i] == ranks[i+1]+1 {
            run++
            if run >= 5 {
                // Found 5 consecutive ranks ending at i+1
                // The high card is ranks[i-3] (since ranks are descending)
                if i >= 3 {
                    return true, ranks[i-3]
                }
            }
        } else if ranks[i] == ranks[i+1] {
            // Duplicate rank, skip
            continue
        } else {
            // Gap in ranks, reset run
            run = 1
        }
    }
    return false, 0
}

func straightFlushHigh(cards []Card) (bool, int) {
    bySuit := suitsMap(cards)
    for _, same := range bySuit {
        if len(same) < 5 {
            continue
        }
        if ok, high := straightHigh(same); ok {
            return true, high
        }
	}
	return false, 0
}

func fourOfAKind(cards []Card) (bool, int, int) {
    counts := rankCounts(cards)
    quadRank := 0
    for r, c := range counts {
        if c == 4 && r > quadRank {
            quadRank = r
        }
    }
    if quadRank == 0 {
        return false, 0, 0
    }
    // Kicker: highest other rank
    kicker := 0
    for r := range counts {
        if r != quadRank && r > kicker {
            kicker = r
        }
    }
    return true, quadRank, kicker
}

func fullHouse(cards []Card) (bool, int, int) {
    counts := rankCounts(cards)
    trips := []int{}
    pairs := []int{}
    for r, c := range counts {
        if c >= 3 {
            trips = append(trips, r)
        } else if c >= 2 {
            pairs = append(pairs, r)
        }
        if c >= 4 {
            // One rank can contribute both a trip and a pair only if there are 5 cards of same rank which is impossible in a single deck.
        }
    }
    sort.Slice(trips, func(i, j int) bool { return trips[i] > trips[j] })
    sort.Slice(pairs, func(i, j int) bool { return pairs[i] > pairs[j] })
    if len(trips) >= 2 {
        return true, trips[0], trips[1]
    }
    if len(trips) >= 1 && len(pairs) >= 1 {
        return true, trips[0], pairs[0]
    }
    return false, 0, 0
}

func flushTop5(cards []Card) (bool, []int) {
    bySuit := suitsMap(cards)
    best := []int{}
    for _, same := range bySuit {
        if len(same) < 5 {
            continue
        }
        sort.Slice(same, func(i, j int) bool { return same[i].Rank > same[j].Rank })
        ranks := []int{}
        for _, c := range same {
            ranks = append(ranks, c.Rank)
            if len(ranks) == 5 {
                break
            }
        }
        if len(ranks) == 5 && (len(best) == 0 || compareIntSlices(ranks, best) > 0) {
            best = ranks
        }
    }
    if len(best) == 5 {
        return true, best
	}
	return false, nil
}

func tripsWithKickers(cards []Card) (bool, int, []int) {
    counts := rankCounts(cards)
    tripRank := 0
    for r, c := range counts {
        if c >= 3 && r > tripRank {
            tripRank = r
        }
    }
    if tripRank == 0 {
        return false, 0, nil
    }
    use := map[int]int{tripRank: 3}
    kickers := topRanksExcluding(cards, use)
    if len(kickers) < 2 {
        return false, 0, nil
    }
    return true, tripRank, kickers[:2]
}

func twoPairWithKicker(cards []Card) (bool, int, int, int) {
    counts := rankCounts(cards)
    pairs := []int{}
    for r, c := range counts {
        if c >= 2 {
            pairs = append(pairs, r)
        }
    }
    sort.Slice(pairs, func(i, j int) bool { return pairs[i] > pairs[j] })
    if len(pairs) < 2 {
        return false, 0, 0, 0
    }
    use := map[int]int{pairs[0]: 2, pairs[1]: 2}
    kicker := topRanksExcluding(cards, use)
    if len(kicker) == 0 {
        return false, 0, 0, 0
    }
    return true, pairs[0], pairs[1], kicker[0]
}

func onePairWithKickers(cards []Card) (bool, int, []int) {
    counts := rankCounts(cards)
    pairRank := 0
    for r, c := range counts {
        if c >= 2 && r > pairRank {
            pairRank = r
        }
    }
    if pairRank == 0 {
        return false, 0, nil
    }
    use := map[int]int{pairRank: 2}
    kickers := topRanksExcluding(cards, use)
    if len(kickers) < 3 {
        return false, 0, nil
    }
    return true, pairRank, kickers[:3]
}

// topRanksExcluding returns ranks sorted desc excluding up to the specified counts per rank.
func topRanksExcluding(cards []Card, exclude map[int]int) []int {
    counts := map[int]int{}
    for r, c := range exclude {
        counts[r] = c
    }
    pool := []int{}
    // Sort cards by rank desc
    cs := append([]Card{}, cards...)
    sort.Slice(cs, func(i, j int) bool { return cs[i].Rank > cs[j].Rank })
    for _, c := range cs {
        if counts[c.Rank] > 0 {
            counts[c.Rank]--
            continue
        }
        // ensure uniqueness by rank for kickers
        if len(pool) == 0 || pool[len(pool)-1] != c.Rank {
            pool = append(pool, c.Rank)
        }
    }
    return pool
}

func compareIntSlices(a, b []int) int {
    n := len(a)
    if len(b) < n {
        n = len(b)
    }
    for i := 0; i < n; i++ {
        if a[i] != b[i] {
            if a[i] > b[i] {
                return 1
            }
            return -1
        }
    }
    if len(a) == len(b) {
        return 0
    }
    if len(a) > len(b) {
        return 1
    }
    return -1
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
