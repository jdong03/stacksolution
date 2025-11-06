package game

import "sort"

func EvaluateHand(flop []Card, playerHand []Card, turn Card, river Card) string {
	return ""
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
