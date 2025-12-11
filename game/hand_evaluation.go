package game

import "sort"

// Hand rank constants (higher is better)
const (
	HighCard      = 1
	OnePair       = 2
	TwoPair       = 3
	ThreeOfAKind  = 4
	Straight      = 5
	Flush         = 6
	FullHouse     = 7
	FourOfAKind   = 8
	StraightFlush = 9
	RoyalFlush    = 10
)

// EvaluateHand evaluates a poker hand and returns:
// - rank: the hand ranking (1-10, higher is better)
// - tiebreakers: ordered list of values for breaking ties (highest first)
func EvaluateHand(flop []Card, playerHand []Card, turn Card, river Card) (int, []int) {
	// Check hands from highest to lowest

	// Royal Flush
	if checkRoyal(flop, playerHand, turn, river) {
		return RoyalFlush, []int{}
	}

	// Straight Flush
	if found, highCard := checkStraightFlush(flop, playerHand, turn, river); found {
		return StraightFlush, []int{highCard}
	}

	// Four of a Kind
	if found, quadRank, kicker := checkQuads(flop, playerHand, turn, river); found {
		return FourOfAKind, []int{quadRank, kicker}
	}

	// Full House
	if found, tripRank, pairRank := checkBoat(flop, playerHand, turn, river); found {
		return FullHouse, []int{tripRank, pairRank}
	}

	// Flush
	if found, _, flushCards := checkFlush(flop, playerHand, turn, river); found {
		tiebreakers := make([]int, 0, 5)
		for i := 0; i < 5 && i < len(flushCards); i++ {
			tiebreakers = append(tiebreakers, flushCards[i].Rank)
		}
		return Flush, tiebreakers
	}

	// Straight
	if found, straightCards := checkStraight(flop, playerHand, turn, river); found {
		return Straight, []int{straightCards[0].Rank}
	}

	// Three of a Kind
	if found, tripRank := checkTrips(flop, playerHand, turn, river); found {
		kickers := getKickers(flop, playerHand, turn, river, []int{tripRank}, 2)
		return ThreeOfAKind, append([]int{tripRank}, kickers...)
	}

	// Two Pair
	if found, highPair, lowPair := checkTwoPair(flop, playerHand, turn, river); found {
		kickers := getKickers(flop, playerHand, turn, river, []int{highPair, lowPair}, 1)
		return TwoPair, append([]int{highPair, lowPair}, kickers...)
	}

	// One Pair
	if found, pairRank := checkPair(flop, playerHand, turn, river); found {
		kickers := getKickers(flop, playerHand, turn, river, []int{pairRank}, 3)
		return OnePair, append([]int{pairRank}, kickers...)
	}

	// High Card
	kickers := getKickers(flop, playerHand, turn, river, []int{}, 5)
	return HighCard, kickers
}

// CompareHands compares two players' hands and returns:
// 1 if player1 wins, -1 if player2 wins, 0 if tie
func CompareHands(p1Cards, p2Cards, flop []Card, turn, river Card) int {
	p1Rank, p1Tiebreakers := EvaluateHand(flop, p1Cards, turn, river)
	p2Rank, p2Tiebreakers := EvaluateHand(flop, p2Cards, turn, river)

	// Higher rank wins
	if p1Rank > p2Rank {
		return 1
	}
	if p2Rank > p1Rank {
		return -1
	}

	// Same rank - compare tiebreakers
	for i := 0; i < len(p1Tiebreakers) && i < len(p2Tiebreakers); i++ {
		if p1Tiebreakers[i] > p2Tiebreakers[i] {
			return 1
		}
		if p2Tiebreakers[i] > p1Tiebreakers[i] {
			return -1
		}
	}

	// Complete tie
	return 0
}

// getKickers returns the top N kickers excluding the specified ranks
func getKickers(flop []Card, playerHand []Card, turn Card, river Card, excludeRanks []int, count int) []int {
	cards := append(playerHand, flop...)
	cards = append(cards, turn, river)

	// Build exclusion map
	exclude := make(map[int]bool)
	for _, r := range excludeRanks {
		exclude[r] = true
	}

	// Get all non-excluded ranks, sorted descending
	ranks := make([]int, 0)
	for _, card := range cards {
		if !exclude[card.Rank] && card.Rank > 0 {
			ranks = append(ranks, card.Rank)
		}
	}

	// Sort descending
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i] > ranks[j]
	})

	// Remove duplicates and take top N
	seen := make(map[int]bool)
	result := make([]int, 0, count)
	for _, r := range ranks {
		if !seen[r] && len(result) < count {
			seen[r] = true
			result = append(result, r)
		}
	}

	return result
}

func checkPair(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	cards := append(playerHand, append(flop, turn, river)...)
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})
	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank {
				count++
			}
		}
		if count == 2 {
			return true, cards[i].Rank
		}
	}
	return false, 0

}

func checkTwoPair(flop []Card, playerHand []Card, turn Card, river Card) (bool, int, int) {

	pairs := []int{}
	usedPairs := map[int]bool{}
	cards := append(playerHand, append(flop, turn, river)...)

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank && !usedPairs[cards[i].Rank] {
				count++
			}
		}
		if count == 2 {
			pairs = append(pairs, cards[i].Rank)
			usedPairs[cards[i].Rank] = true
		}
	}

	if len(pairs) >= 2 {
		return true, pairs[0], pairs[1]
	}
	return false, 0, 0

}

func checkTrips(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	trips := []int{}
	usedTrips := map[int]bool{}
	cards := append(playerHand, append(flop, turn, river)...)

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank && !usedTrips[cards[i].Rank] {
				count++
			}
		}
		if count == 3 {
			trips = append(trips, cards[i].Rank)
			usedTrips[cards[i].Rank] = true
		}
	}

	if len(trips) != 1 {
		return false, 0
	}

	if len(trips) == 1 {
		return true, trips[0]
	}
	return false, 0
}

func checkStraight(flop []Card, playerHand []Card, turn Card, river Card) (bool, []Card) {
	cards := append(playerHand, append(flop, turn, river)...)
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	nonPair := []Card{}
	for i := 0; i < len(cards); i++ {
		if i == 0 || cards[i].Rank != cards[i-1].Rank {
			nonPair = append(nonPair, cards[i])
			if cards[i].Rank == 14 {
				nonPair = append(nonPair, Card{1, cards[i].Suit})
			}
		}
	}

	sort.Slice(nonPair, func(i, j int) bool {
		return nonPair[i].Rank > nonPair[j].Rank
	})

	var straightCards []Card
	var count = 1
	for i := 0; i < len(nonPair)-1; i++ {
		if len(straightCards) == 0 {
			straightCards = append(straightCards, nonPair[i])
		}
		if nonPair[i].Rank == nonPair[i+1].Rank+1 {
			count++
			straightCards = append(straightCards, nonPair[i+1])
		} else {
			straightCards = []Card{}
			count = 1
		}
		if count >= 5 {
			return true, straightCards
		}

	}
	return false, nil

}

func checkFlush(flop []Card, playerHand []Card, turn Card, river Card) (bool, string, []Card) {
	cards := append(playerHand, append(flop, turn, river)...)

	spades := 0
	hearts := 0
	diamonds := 0
	clubs := 0

	for i := 0; i < len(cards); i++ {
		suit := cards[i].Suit

		switch suit {
		case "Spades":
			spades++
		case "Hearts":
			hearts++
		case "Diamonds":
			diamonds++
		case "Clubs":
			clubs++
		}
	}

	var flushSuit string
	if spades >= 5 {
		flushSuit = "Spades"
	} else if hearts >= 5 {
		flushSuit = "Hearts"
	} else if diamonds >= 5 {
		flushSuit = "Diamonds"
	} else if clubs >= 5 {
		flushSuit = "Clubs"
	} else {
		return false, "", nil
	}

	flushCards := []Card{}
	for i := 0; i < len(cards); i++ {
		if cards[i].Suit == flushSuit {
			flushCards = append(flushCards, cards[i])
		}
	}

	sort.Slice(flushCards, func(i, j int) bool {
		return flushCards[i].Rank > flushCards[j].Rank
	})

	return true, flushSuit, flushCards

}

/*
checkBoat checks for full house
return (bool found, int tripRank, int pairRank)
*/
func checkBoat(flop []Card, playerHand []Card, turn Card, river Card) (bool, int, int) {
	// Count frequency of each rank
	cards := append(playerHand, append(flop, turn, river)...)
	rankFreq := getRankFrequencyMap(cards)

	// Look for three of a kind (freq >= 3) and a pair (freq == 2)
	var tripRank, pairRank int
	for rank, freq := range rankFreq {
		if freq >= 3 {
			if rank > tripRank {
				pairRank = tripRank
				tripRank = rank
			} else {
				if rank > pairRank {
					pairRank = rank
				}
			}
		} else if freq == 2 {
			if rank > pairRank {
				pairRank = rank
			}
		}
	}

	// Check if both three of a kind and pair were found
	if tripRank > 0 && pairRank > 0 {
		return true, tripRank, pairRank
	}
	return false, 0, 0
}

/*
checkQuads checks for four of a kind
return (bool found, int quadRank, int kicker)
*/
func checkQuads(flop []Card, playerHand []Card, turn Card, river Card) (bool, int, int) {
	// Count frequency of each rank
	cards := append(playerHand, append(flop, turn, river)...)
	rankFreq := getRankFrequencyMap(cards)

	// Look for four of a kind (freq == 4) and get the highest kicker
	var quadRank, kicker int
	for rank, freq := range rankFreq {
		if freq == 4 {
			quadRank = rank
		} else if freq > 0 && rank > kicker {
			kicker = rank
		}
	}

	// Check if four of a kind was found
	if quadRank > 0 {
		return true, quadRank, kicker
	}
	return false, 0, 0
}

/*
checkStraighFlush checks for straight flush
return (bool found, int highCardRank(top of straight flush) )
*/
func checkStraightFlush(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	// Bucket cards by suit
	cards := append(playerHand, append(flop, turn, river)...)
	suitBuckets := map[string][]Card{}
	for _, card := range cards {
		suitBuckets[card.Suit] = append(suitBuckets[card.Suit], card)
	}

	// Check each suit bucket for straight
	for _, suitedCards := range suitBuckets {
		if len(suitedCards) >= 5 {
			sort.Slice(suitedCards, func(i, j int) bool {
				return suitedCards[i].Rank > suitedCards[j].Rank
			})

			// Add Ace as low card if present
			if suitedCards[len(suitedCards)-1].Rank == 14 {
				suitedCards = append(suitedCards, Card{Rank: 1, Suit: suitedCards[0].Suit})
			}

			// Look for straight in suited cards
			var count = 1
			for i := 0; i < len(suitedCards)-1; i++ {
				if suitedCards[i].Rank == suitedCards[i+1].Rank+1 {
					count++
				} else {
					count = 1
				}

				if count >= 5 {
					return true, suitedCards[i-3].Rank
				}
			}
		}
	}
	return false, 0
}

/*
checkRoyal checks for royal flush
return bool found
*/
func checkRoyal(flop []Card, playerHand []Card, turn Card, river Card) bool {
	// TODO: Aaron N should implement
	return false
}

// ============= Helper Functions =============

/*
Helper function to get rank frequency map from a set of cards
*/
func getRankFrequencyMap(cards []Card) map[int]int {
	rankFreq := make(map[int]int)
	for _, card := range cards {
		rankFreq[card.Rank]++
	}
	return rankFreq
}
