package game

// Player represents a player in the game with a hand of cards and a stack of chips.
type Player struct {
	Hand  []Card
	Stack int
}

// NewPlayer creates and returns a new player with an initial stack.
func NewPlayer(initialStack int) *Player {
	return &Player{
		Hand:  []Card{},
		Stack: initialStack,
	}
}

// addToStack adds the specified amount to the player's stack.
func (p *Player) addToStack(amount int) {
	p.Stack += amount
}

// removeFromStack removes the specified amount from the player's stack.
func (p *Player) removeFromStack(amount int) {
	p.Stack -= amount
}
