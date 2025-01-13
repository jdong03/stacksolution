package game

import "sort"

func evaluateHand(flop []Card, playerHand []Card, turn Card, river Card) string {
}

func checkPair(flop []Card, playerHand []Card, turn Card, river Card) (bool, int) {
	cards := append(playerHand, append(flop, turn, river)...)

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

func checkTrips(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkStraight(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkFlush(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkBoat(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkQuads(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkSF(flop []Card, playerHand []Card, turn Card, river Card) bool {

}

func checkRoyal(flop []Card, playerHand []Card, turn Card, river Card) bool {

}
