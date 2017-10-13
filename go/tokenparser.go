package parsejoy

import "fmt"
import "strings"
import "7scientists.com/parsejoy/set"

type TokenState struct {
	CurrentToken *Token
	Context *TokenContext
    //The AST stack
    ASTStack []interface{}

	Level int
	Debug bool
}

type TokenContext struct {
	Errors     uint
	Calls      uint
    StringState *StringState
}

type L2Token struct {
	Type     string
	Ignore   bool
	From *Token
	To *Token
	Children *L2Token
	Parent *L2Token
	Next *L2Token
	Last *L2Token
	First *L2Token
}

func (self *L2Token) Copy() *L2Token {
	newToken := *self
	return &newToken
}

func (self *L2Token) GetValue(state State) (s string) {
    tokenState, ok := state.(*TokenState)
    if !ok{
        return
    }
    if self.From != nil && self.To != nil {
        return string(tokenState.Context.StringState.Context.S[self.From.From.Position:self.To.From.Position])
    }
    return
}

func (self *L2Token) SetNext(token BaseToken) {
	if token == nil {
		currentToken := self
		currentToken.Parent = nil
		for currentToken.Next != nil {
			nextToken := currentToken.Next
			currentToken.Next = nil
			currentToken.Parent = nil
			currentToken.Last = nil
			currentToken = nextToken
		}
		return
	}
	l2Token, ok := token.(*L2Token)
	if ! ok {
		panic("Expected L2Token")
	}

	currentToken := self
	for currentToken != nil && currentToken.Next != nil {
		currentToken = currentToken.Next
	}
	currentToken.Next = l2Token
	l2Token.First = self
	self.Last = l2Token
}

func (s *TokenState) PushASTNode(node *ASTNode, oldState State) error {
    //Pushes an AST node to the stack.
    //This will go through the current stack (which should contain only properties)
    //and add all the properties to the AST node.
    //Empties the stack afterwards
    oldTokenState, ok := oldState.(*TokenState)
    if !ok {
        return &ParserError{"Expected a token state!"}
    }
    oldLength := len(oldTokenState.ASTStack)
    oldStack := s.ASTStack
    newASTProperties := s.ASTStack[oldLength:]
    for _, stackElement := range newASTProperties {

        astProperty, ok := stackElement.(*ASTProperty)
        if !ok {
            return &ParserError{"Expected an AST property"}
        }
        node.Attributes[astProperty.Name] = astProperty.Value
    }
    //we replace the current stack by the new AST node
    s.ASTStack = append(oldStack[:oldLength], node)
    return nil
}

func (s *TokenState) PushASTProperty(property *ASTProperty, oldState State) error {
    //Pushes an AST property to the stack.
    //Takes the current stack (which should contain only ASTNodes) and adds them
    //to the value of the given property.
    oldTokenState, ok := oldState.(*TokenState)
    if !ok {
        return &ParserError{"Expected a token state!"}
    }
    oldLength := len(oldTokenState.ASTStack)
    newASTNodes := make([]interface{},len(s.ASTStack[oldLength:]))
    copy(newASTNodes,s.ASTStack[oldLength:])
    if property.Value == nil {
        property.Value = newASTNodes
    }
    s.ASTStack = append(s.ASTStack[:oldLength], property)
    return nil
}


func (s *TokenState) CreateToken(_type string, from *Token, to *Token) *L2Token {

	token := new(L2Token)
	token.Type = _type
	token.From = from
	token.To = to
	token.Ignore = false
	token.First = token
	token.Last = token

	if len(_type) >= 2 && _type[:2] == "__" {
		token.Ignore = true
	}
	return token
}

func (self *TokenState) Advance(token *Token) {
	self.CurrentToken = token
	if self.CurrentToken != nil {
		self.CurrentToken = self.CurrentToken.Next
	}
	for self.CurrentToken != nil && self.CurrentToken.Ignore {
		self.CurrentToken = self.CurrentToken.Next
	}
}

func (self *TokenState) Get(tokenId uint, leavesOnly bool) *Token {

	currentToken := self.CurrentToken
	if currentToken == nil {
		return nil
	}
	for currentToken != nil && currentToken.Ignore {
		currentToken = currentToken.Next
	}
	if tokenId > 0 {
		for currentToken.Id != tokenId {
			if currentToken.Children != nil {
				currentToken = currentToken.Children
				for currentToken != nil && currentToken.Ignore {
					currentToken = currentToken.Next
				}
			} else {
				return nil
			}
		}
	} else if leavesOnly {
		for currentToken.Children != nil {
			currentToken = currentToken.Children
			for currentToken != nil && currentToken.Ignore {
				currentToken = currentToken.Next
			}
		}
	}
	return currentToken
}

func (self *TokenState) Copy() State {
	newState := *self
	return &newState
}

func LinkTokens(token *Token) {
	currentToken := token

	if currentToken.Children != nil {
		LinkTokens(currentToken.Children)
	}

	for currentToken.Next != nil {
		currentToken = currentToken.Next
		if currentToken.Children != nil {
			LinkTokens(currentToken.Children)
		}
	}

	currentParent := currentToken.Parent
	for currentParent != nil {
		if currentParent.Next != nil {
			currentToken.Next = currentParent.Next
			break
		}
		currentParent = currentParent.Parent
	}

}

func (s *TokenState) Initialize(token *Token, stringState *StringState) {
	s.CurrentToken = token
    s.ASTStack = make([]interface{}, 0, 1)
	s.Context = new(TokenContext)
    s.Context.StringState = stringState
}

func PrintL2ParseTree(token *L2Token, level int) int {
	nodes := 0
	currentToken := token
	for currentToken != nil {
		nodes += 1
		fmt.Println(strings.Repeat(" ", level), currentToken.Type)
		if currentToken.Children != nil {
			nodes += PrintL2ParseTree(currentToken.Children, level+1)
		}
		currentToken = currentToken.Next
	}
	if level == 0 {
		fmt.Println("Nodes:", nodes)
	}
	return nodes
}

type TokenParserPlugin struct {
	ParserPluginMixin
	tokenGrammar     *set.BitGrammar
	nilId        uint
	falseId      uint
}

func (pg *TokenParserPlugin) PrepareGrammar(grammar map[string]interface{}) {

}

func (pg *TokenParserPlugin) Initialize(tokenIds *set.BitGrammar) {
	pg.ParserPluginMixin.Initialize()
	pg.tokenGrammar = tokenIds
	pg.nilId = pg.tokenGrammar.GetOrAdd(nil)
	pg.falseId = pg.tokenGrammar.GetOrAdd(false)
}

func (pg *TokenParserPlugin) canProceed(prefixes *set.BitSet, state *TokenState) bool {

	currentToken := state.CurrentToken

	for currentToken != nil && currentToken.Ignore {
		currentToken = currentToken.Next
	}

	if currentToken == nil {
		return false
	}

	for {
		if prefixes.ContainsId(currentToken.Id) {
			return true
		}
		if currentToken.Children != nil {
			currentToken = currentToken.Children
			for currentToken != nil && currentToken.Ignore {
				currentToken = currentToken.Next
			}
		} else {
			return false
		}
	}
	return false
}

func (pg *TokenParserPlugin) getEmptyPrefix(args interface{}) set.Set {
	grammar, ok := args.(*set.BitGrammar)
	if !ok {
		panic("Expected bit grammar!")
	}
	prefixes := set.BitSet{}
	prefixes.Initialize(grammar)
	return &prefixes
}

func (pg *TokenParserPlugin) getPrefix(rule interface{}, args interface{}) set.Set {

	grammar, ok := args.(*set.BitGrammar)
	if !ok {
		panic("Expected bit grammar!")
	}
	prefixes := set.BitSet{}
	prefixes.Initialize(grammar)

	var stringRule string

	dictRule, ok := rule.(map[interface{}]interface{})
	if !ok {
		stringRule, ok = rule.(string)
		if !ok {
			fmt.Println(rule)
			panic("uh oh!!!")
		}
	} else {
		fmt.Println(dictRule)
		tokenRule, ok := dictRule["token"]
		if !ok {
			panic("uh oh")
		}
		stringRule, ok = tokenRule.(string)
		if ! ok {
			panic("uh oh!")
		}
	}
	if string(stringRule[len(stringRule)-1]) == "!" {
		prefixes.Add(string(stringRule[0 : len(stringRule)-1]))
	} else {
		prefixes.Add(stringRule)
	}
	return &prefixes
}

func (pg *TokenParserPlugin) parser(parser Parser, name string, rule interface{}, emit bool) Parser {

	if len(name) >= 2 && name[0:2] == "__"{
		emit = false
		//return parser
	}
	prefixes := pg.generator.getPrefixes(rule, set.HashSet{make(map[interface{}]bool)}, pg.tokenGrammar, true)
	bitPrefixes, ok := prefixes.(*set.BitSet)
	if !ok {
		panic("Expected bit prefixes!")
	}

	usePrefixes := true
	storeResult := false
	if name == "or_test" {
		storeResult = true
	}
	if bitPrefixes.ContainsId(pg.nilId) || prefixes.N() == 0 || bitPrefixes.ContainsId(pg.falseId) {
		usePrefixes = false
	}
	fingerprint := pg.generator.GetFingerprint(rule)

	fingerprintId, ok := pg.fingerprintIds[fingerprint]

	if !ok {
		fingerprintId = pg.fingerprintId
		pg.fingerprintIds[fingerprint] = pg.fingerprintId
		pg.fingerprintId += 1
	}

	pe:= &ParserError{"Cannot proceed with this token"}

	wrappedParser := func(state State) (State, BaseToken, error) {
		tokenState, ok := state.(*TokenState)

		if !ok {
			return state, nil, &ParserError{"Invalid state: Expceted a token state in wrapper"}
		}

		currentToken := tokenState.CurrentToken

		level := tokenState.Level

		if usePrefixes && !pg.canProceed(bitPrefixes, tokenState) {
			if tokenState.Debug {
				tokenIds := make([]string,0,10)
				ct := currentToken
				for ct != nil {
					tokenId := pg.tokenGrammar.ValueForId(ct.Id)
					stringTokenId, _ := tokenId.(string)
					tokenIds = append(tokenIds,stringTokenId)
					ct = ct.Children
				}
				fmt.Println(strings.Repeat(" ",tokenState.Level),"cannot proceed with",name,"(token:",strings.Join(tokenIds,","),")")
			}
			return state, nil, pe
		}

		if tokenState.Debug {
			fmt.Println(strings.Repeat(" ",level),"Entering",name,currentToken.From.Row,":",currentToken.From.Column,"-",currentToken.To.Row,":",currentToken.To.Column)
		}


		if storeResult && currentToken.Outcomes != nil {
			var i int
			for i = range currentToken.Outcomes {
				j := len(currentToken.Outcomes)-i-1
				if currentToken.Outcomes[j].Fingerprint == fingerprintId {
					//fmt.Println("Restoring:",name,currentToken.Id)
					return currentToken.Outcomes[j].State, currentToken.Outcomes[j].Token, currentToken.Outcomes[j].Error
				}
			}
		}

		tokenState.Level += 1
		newState, token, err := parser(state)
		l2Token, ok := token.(*L2Token)
		newTokenState, ok := newState.(*TokenState)
		if !ok {
			return state, nil, &ParserError{"Expected token state"}
		}

		newTokenState.Level = level
		tokenState.Level = level

		var newToken BaseToken

		if err != nil {
			if tokenState.Debug {
				fmt.Println(strings.Repeat(" ",level),"Error for",name,":",err)
			}
			newTokenState = tokenState
		} else {
			if tokenState.Debug {
				tokenId := pg.tokenGrammar.ValueForId(currentToken.Id)
				stringTokenId, _ := tokenId.(string)
				fmt.Println(strings.Repeat(" ",level),"Success for",name,stringTokenId)
			}

			if emit {
				newL2Token := newTokenState.CreateToken(name,tokenState.CurrentToken,newTokenState.CurrentToken)
				if l2Token != nil {
					newL2Token.Children = l2Token
					l2Token.Parent = newL2Token
				}
				newToken = newL2Token
			} else {
				newToken = token
			}
		}

		if storeResult{
			if currentToken.Outcomes == nil {
				currentToken.Outcomes = make([]Outcome,0,5)
			}
			var newL2Token *L2Token
			if newToken != nil {
				newL2Token, _ = newToken.(*L2Token)
			}
			//fmt.Println("Storing:",name,currentToken.Id)
			currentToken.Outcomes = append(currentToken.Outcomes,Outcome{newTokenState, fingerprintId, newL2Token, err})
		}

		newTokenState.Context.Calls += 1
		return newTokenState, newToken, err
	}
	return wrappedParser
}

func (pg *TokenParserPlugin) resolveRule(generator *ParserGenerator, name string) (func(rule interface{}) (Parser, error), bool) {
	switch name {
	default:
		tokenParser := func(rule interface{}) (Parser, error) {
			if name[len(name)-1] == '!' {
				name = name[0 : len(name)-1]
			}
			tokenRule := make(map[interface{}]interface{})
			tokenRule["token"] = name
			return pg.compileToken(tokenRule)
		}
		return tokenParser, true
	}
	return nil, false
}

func (pg *TokenParserPlugin) compileToken(rule interface{}) (Parser, error) {

	ruleDict, ok := rule.(map[interface{}]interface{})

	if !ok {
		return nil, &CompilerError{"Expected a dictionary"}
	}

	tokenRule, ok := ruleDict["token"]

	if !ok {
		return nil, &CompilerError{"token not found"}
	}

	tokenType, ok := tokenRule.(string)

	if !ok {
		return nil, &CompilerError{"Token not a string"}
	}

	tokenId := pg.tokenGrammar.GetOrAdd(tokenType)
	pe := &ParserError{fmt.Sprintf("Expected a token of type %s",tokenType)}

	tokenParser := func(state State) (State, BaseToken, error) {

		tokenState, ok := state.(*TokenState)
		if !ok {
			return state, nil, &ParserError{"Invalid state!!!!"}
		}
		currentToken := tokenState.Get(tokenId, false)
		if currentToken == nil {
			return tokenState, nil, pe
		}
		newState := tokenState.Copy()
		newTokenState, _ := newState.(*TokenState)
		newTokenState.Advance(currentToken)
		return newTokenState, nil, nil

	}

	return pg.generator.parser(tokenParser, tokenType, rule, true), nil
}
