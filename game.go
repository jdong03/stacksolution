package game

import "sort"

func evaluateHand(flop []Card, playerHand []Card, turn Card, river Card) string {
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

func checkBoat(flop []Card, playerHand []Card, turn Card, river Card) (bool, int, int) {
	//return bool, trip, pair
	trips := []int{}
	usedTrips := map[int]bool{}
	pairs := []int{}
	usedPairs := map[int]bool{}
	cards := append(playerHand, append(flop, turn, river)...)
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank > cards[j].Rank
	})

	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank && !usedTrips[cards[i].Rank] && !usedPairs[cards[i].Rank] {
				count++
			}
		}
		if count == 3 {
			trips = append(trips, cards[i].Rank)
			usedTrips[cards[i].Rank] = true
		}
		if count == 2 {
			pairs = append(pairs, cards[i].Rank)
			usedPairs[cards[i].Rank] = true
		}
	}

	if len(trips) > 1 {
		return true, trips[0], trips[1]
	}
	if len(trips) == 1 && len(pairs) > 0 {
		return true, trips[0], pairs[0]
	} else {
		return false, 0, 0
	}

}

func checkQuads(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	cards := append(playerHand, append(flop, turn, river)...)
	for i := 0; i < len(cards); i++ {
		count := 0
		for j := 0; j < len(cards); j++ {
			if cards[i].Rank == cards[j].Rank {
				count++
			}
		}
		if count == 4 {
			return true, cards[i].Rank
		}
	}
	return false, 0
}

func checkSF(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkRoyal(flop []Card, playerHand []Card, turn Card, river Card) bool {

}
