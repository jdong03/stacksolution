package game

func testCheckPair() bool {
	playerHand := []Card{
		{Rank: 10, Suit: "Hearts"},
		{Rank: 10, Suit: "Clubs"},
	}
	flop := []Card{
		{Rank: 3, Suit: "Diamonds"},
		{Rank: 7, Suit: "Spades"},
		{Rank: 2, Suit: "Hearts"},
	}
	turn := Card{Rank: 9, Suit: "Clubs"}
	river := Card{Rank: 4, Suit: "Diamonds"}

	hasPair, pairRank := checkPair(flop, playerHand, turn, river)

	if hasPair == true && pairRank == 10 {
		return true
	} else {
		return false
	}
}
