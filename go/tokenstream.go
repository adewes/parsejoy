package parsejoy

type Position struct {
	Position int
	Column int
	Row int
}

type Token struct {
	Id       uint
	Number uint
	Ignore   bool
	From     Position
	To       Position
	Next *Token
	Last *Token
	Children *Token
	Parent *Token
	Outcomes []Outcome
}

type Outcome struct {
	State       State
	Fingerprint uint
	Token BaseToken
	Error       error
}

func (self *Token) GetValue(state State) (s string) {
    stringState, ok := state.(*StringState)
    if !ok {
        return ""
    }
    return string(stringState.Context.S[self.From.Position:self.To.Position])
}

func (self *Token) SetNext(token BaseToken) {
	if token == nil {
		currentToken := self
		for currentToken.Next != nil {
			nextToken := currentToken.Next
			currentToken.Next = nil
			currentToken.Parent = nil
			currentToken = nextToken
		}
		return
	}
	stringToken, ok := token.(*Token)
	if !ok {
		panic("Expected a token!")
	}
	currentToken := self
	for currentToken.Next != nil {
		currentToken = currentToken.Next
	}
	currentToken.Next = stringToken
	stringToken.Parent = self.Parent
	stringToken.Last = stringToken
}
