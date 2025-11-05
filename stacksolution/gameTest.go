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

	rank := Evaluate7(flop, playerHand, turn, river)
	return rank.Category == OnePair && rank.Primary[0] == 10
}

func testCheckTwoPair() bool {
	playerHand := []Card{
		{Rank: 3, Suit: "Hearts"},
		{Rank: 10, Suit: "Clubs"},
	}
	flop := []Card{
		{Rank: 3, Suit: "Diamonds"},
		{Rank: 10, Suit: "Spades"},
		{Rank: 2, Suit: "Hearts"},
	}
	turn := Card{Rank: 9, Suit: "Clubs"}
	river := Card{Rank: 4, Suit: "Diamonds"}

	rank := Evaluate7(flop, playerHand, turn, river)
	return rank.Category == TwoPair && rank.Primary[0] == 10 && rank.Primary[1] == 3
}

func testCheckTrips() bool {
	// TODO: Add test case for three of a kind
	return false
}
