package parsejoy

import "fmt"
import "regexp"
import "strings"
import "bytes"
import "7scientists.com/parsejoy/set"

type StringParserPlugin struct {
	ParserPluginMixin
	PrefixGrammar *set.BitGrammar //contains all possible prefixes
    PrefixList [][]byte
    PrefixIds []uint
	TokenIds *set.BitGrammar
	PrefixTree *Prefix
    MaxPrefixLength int
}

type Prefix struct {
    Parent *Prefix
    Value []byte
    Id uint
    Next []*Prefix
}

type StringContext struct {
	TokenNumber uint
    TokenIds *set.BitGrammar
	TokenId  uint
	Errors     uint
	Calls      uint
	S          []byte
	CurrentRow int
	LineBreaks []int
	NumberOfLines int
}

type StringState struct {
	Pos     int
	N       int
	Level int
	Debug bool
	CurrentPrefixSet *PrefixSet
	Indents [][]byte
	Context *StringContext
}

type PrefixSet struct {
    Set *set.BitSet
    Pos int
}

func (s *StringState) CreateToken(tokenId uint, from int, to int, ignore bool) (*Token, bool) {
	token := new(Token)
	token.Id = tokenId
	token.Number = s.Context.TokenNumber
	s.Context.TokenNumber+=1
	token.Ignore = ignore
	token.From = s.GetPosition(from)
	token.To = s.GetPosition(to)
    return token, true
}

func (s *StringState) PushASTNode(node *ASTNode, state State) error {
    return nil
}

func (s *StringState) PushASTProperty(node *ASTProperty, state State) error {
    return nil
}

func (s *StringState) GetPosition(position int) Position {
	var p Position
	if len(s.Context.LineBreaks) == 0 {
		return p
	}
	for s.Context.LineBreaks[s.Context.CurrentRow] < position {
		if s.Context.CurrentRow >= len(s.Context.LineBreaks)-1 {
			break
		}
		s.Context.CurrentRow += 1
	}
	for s.Context.LineBreaks[s.Context.CurrentRow] > position {
		if s.Context.CurrentRow > 0 {
			if s.Context.LineBreaks[s.Context.CurrentRow-1] <= position {
				break
			} else {
				s.Context.CurrentRow -= 1
			}
		} else {
			break
		}
	}
	p.Position = position
	p.Row = s.Context.CurrentRow
	p.Column = s.Context.LineBreaks[s.Context.CurrentRow]-position
	return p
}

func PrintParseTree(token *Token, grammar *set.BitGrammar, level int) {
	currentToken := token
	for currentToken != nil {
		var nextNumber,parentNumber uint

		if currentToken.Next != nil {
			nextNumber = currentToken.Next.Number
		}
		if currentToken.Parent != nil {
			parentNumber = currentToken.Parent.Number
		}
		fmt.Println(strings.Repeat(" ", level), grammar.ValueForId(currentToken.Id),"(id:",currentToken.Id,", number:",currentToken.Number,")","ignore:",currentToken.Ignore,"from:",currentToken.From.Position,"to:",currentToken.To.Position,"next:",nextNumber,"parent:",parentNumber)
		if currentToken.Children != nil {
			PrintParseTree(currentToken.Children, grammar, level+1)
		}
		if currentToken.Next != nil && currentToken.Next.Parent != currentToken.Parent {
			break
		}
		currentToken = currentToken.Next

	}
}

func (s *StringState) Initialize(code []byte, tokenIds *set.BitGrammar) {
	s.N = len(code)
	s.Indents = make([][]byte,1,10)
	s.Indents[0] = make([]byte,0,4)
	s.Context = &StringContext{}
	s.Context.Initialize(uint(s.N) + 1, tokenIds)
	s.Context.S = code
	s.Context.LineBreaks = make([]int,0,100)
	byteString := []byte("\n")
	for i := range code {
		if code[i] == byteString[0] {
		  s.Context.LineBreaks = append(s.Context.LineBreaks,int(i))
		}
	}
	s.Context.NumberOfLines = len(s.Context.LineBreaks)
}

func (s *StringContext) Initialize(N uint, tokenIds *set.BitGrammar) {
	s.TokenId = 1
    s.TokenIds = tokenIds
}

func (s *StringState) Copy() State {
	newS := *s
	newS.Indents = make([][]byte,len(s.Indents))
	copy(newS.Indents,s.Indents)
	return &newS
}

func (s *StringState) Value() []byte {
	if s.Pos >= s.N {
		return make([]byte, 0)
	}
	return s.Context.S[s.Pos:]
}

func (s *StringState) HasPrefix(prefix []byte) bool {
	return bytes.Compare(s.Context.S[s.Pos:s.Pos+len(prefix)], prefix) == 0
}

func (s *StringState) ValueN(n int) []byte {
	if s.Pos >= s.N {
		return make([]byte, 0)
	}
	i := s.N
	if s.Pos+n < s.N {
		i = s.Pos + n
	}
	return s.Context.S[s.Pos:i]
}

func (s *StringState) Advance(n int) bool {
    if n+s.Pos <= s.N {
		s.Pos += n
		return true
	} else {
		s.Pos = s.N
	}
	return false
}

func buildPrefixTree(values map[interface{}]uint) *Prefix {
	prefix := new(Prefix)
	prefix.Next = make([]*Prefix,256)

    for key, id := range values {
		stringKey, ok := key.(string)
		if !ok {
			continue
		}
		currentPrefix := prefix
		byteKey := []byte(stringKey)
		for i := 0 ; i < len(byteKey) ; i++ {
	        b := byteKey[i]
			newPrefix := currentPrefix.Next[b]
			if newPrefix == nil {
				newPrefix = new(Prefix)
				newPrefix.Next = make([]*Prefix,256)
				currentPrefix.Next[b] = newPrefix
			}
			currentPrefix = newPrefix
			if i == len(byteKey) - 1 {
				newPrefix.Id = id
				newPrefix.Value = byteKey
			}
		}
	}
	return prefix
}

func (pg *StringParserPlugin) PrepareGrammar(grammar map[string]interface{}) {
	pg.PrefixGrammar = new(set.BitGrammar)
    pg.PrefixGrammar.Initialize()
    pg.PrefixList = make([][]byte,0,100)
    pg.PrefixIds = make([]uint,0,100)
    prefixes := pg.generator.getPrefixes(grammar["start"],set.HashSet{make(map[interface{}]bool)}, nil, false)
    hashPrefixes ,ok := prefixes.(*set.HashSet)
    if !ok{
        panic("...")
    }
    prefixList := hashPrefixes.AsList()
    for i := range prefixList {
        prefix := prefixList[i]
        stringPrefix, ok := prefix.(string)
        if ok {
            id := pg.PrefixGrammar.GetOrAdd(stringPrefix)
            pg.PrefixList = append(pg.PrefixList,[]byte(stringPrefix))
            pg.PrefixIds = append(pg.PrefixIds,id)
            if len(stringPrefix) > pg.MaxPrefixLength {
                pg.MaxPrefixLength = len(stringPrefix)
            }
        }
    }
	pg.PrefixTree = buildPrefixTree(pg.PrefixGrammar.Mapping)
}

func (pg *StringParserPlugin) updateCurrentPrefixSet(state *StringState) {
	p := state.Pos
	if state.CurrentPrefixSet == nil {
		state.CurrentPrefixSet = new(PrefixSet)
		state.CurrentPrefixSet.Set = new(set.BitSet)
		state.CurrentPrefixSet.Set.Initialize(pg.PrefixGrammar)
	}
	state.CurrentPrefixSet.Set.Reset()
	state.CurrentPrefixSet.Pos = p
	currentPrefix := pg.PrefixTree
	for {
		if p >= state.N {
			if state.Debug {
				fmt.Println(strings.Repeat(" ",state.Level),"At the end of the input!")
			}
			break
		}
		b := state.Context.S[p]
		nextPrefix := currentPrefix.Next[b]
		if nextPrefix == nil {
			break
		}
		if nextPrefix.Id > 0 {
			if state.Debug {
				fmt.Println(strings.Repeat(" ",state.Level),"prefix:'",string(b),"'")
			}
			state.CurrentPrefixSet.Set.AddById(nextPrefix.Id)
		}
		currentPrefix = nextPrefix
		p += 1
	}
}

func (pg *StringParserPlugin) Initialize() {
	pg.ParserPluginMixin.Initialize()
	pg.TokenIds = new(set.BitGrammar)
	pg.TokenIds.Initialize()
}

func (pg *StringParserPlugin) getEmptyPrefix(args interface{}) set.Set {
	prefixes := set.HashSet{}
	prefixes.Initialize()
	return &prefixes
}

func (pg *StringParserPlugin) getPrefix(rule interface{}, args interface{}) set.Set {

    nilPrefix := new(set.HashSet)
    nilPrefix.Initialize()
    nilPrefix.Add(nil)
	prefixes := new(set.HashSet)
	prefixes.Initialize()
	stringRule, ok := rule.(string)
	if ok {
		if stringRule == "$indent" || stringRule == "$eof" {
			return nilPrefix
		}
		prefixes.Add(stringRule)
		return prefixes
	}
	dictRule, ok := rule.(map[interface{}]interface{})
	if ok {
		var key, value interface{}
		for key, value = range dictRule {
			break
		}
		stringKey, ok := key.(string)
		if !ok {
			return nilPrefix
		}
		stringValue, ok := value.(string)
		if !ok {
			return nilPrefix
		}
		if stringKey == "$literal" {
			prefixes.Add(stringValue)
			return prefixes
		}
	}
    return nilPrefix
}

func (pg *StringParserPlugin) compileIndent(rule interface{}) (Parser, error) {

	indentRule, ok := rule.(string)

	if !ok {
		return nil, &CompilerError{"Expected a single string"}
	}

	if indentRule != "$indent" {
		return nil, &CompilerError{"Invalid rule name!"}
	}

	indentId := pg.TokenIds.GetOrAdd("indent")
	dedentId := pg.TokenIds.GetOrAdd("dedent")
	currentIndentId := pg.TokenIds.GetOrAdd("current_indent")

	indentParser := func(state State) (State, BaseToken, error) {

		stringState, ok := state.(*StringState)
		if !ok {
			return state, nil, &ParserError{"Invalid state: Expected a string state"}
		}

		var indents [][]byte = stringState.Indents

		p := stringState.Pos
		for {
			if p >= stringState.N || (stringState.Context.S[p] != byte("\t"[0]) && stringState.Context.S[p] != byte(" "[0])) {
				break
			}
			p++
		}
		indentString := stringState.Context.S[stringState.Pos:p]
		currentIndent := indents[len(indents)-1]
		newStringState := stringState
		var baseToken BaseToken

		if bytes.Compare(indentString, currentIndent) == 0 {
			newToken, _ := stringState.CreateToken(currentIndentId, p, p+len(currentIndent), true)
			baseToken = newToken
			newStringState.Advance(len(currentIndent))
		} else if len(indentString) > len(currentIndent) && bytes.Compare(indentString[:len(currentIndent)], currentIndent) == 0 {
			newState := stringState.Copy()
			newStringState, ok = newState.(*StringState)
			newToken, _ := stringState.CreateToken(currentIndentId, p, p+len(currentIndent), true)
			baseToken = newToken
			newStringState.Advance(len(currentIndent))
			newIndent := indentString[len(currentIndent):len(indentString)]
			newToken, _ = stringState.CreateToken(indentId, p, p+len(newIndent), false)
			baseToken.SetNext(newToken)
			newStringState.Advance(len(newIndent))

			//we copy the indentation array
			newStringState.Indents = append(indents, indentString)
		} else {
			newState := stringState.Copy()
			newStringState, ok = newState.(*StringState)
			possibleIndents := indents[0 : len(indents)-1]
			cnt := 0
			i := len(possibleIndents) - 1

		Loop:
			for {
				if i < 0 {
					return stringState, nil, &ParserError{"Dedentation does not match!"}
				}
				cnt += 1
				possibleIndent := possibleIndents[i]
				if bytes.Compare(possibleIndent, indentString) == 0 {
					for j := 0; j < cnt; j += 1 {
						newToken, _ := stringState.CreateToken(dedentId, p, p, false)
						if baseToken == nil {
							baseToken = newToken
						} else {
							baseToken.SetNext(newToken)
						}
					}
					newToken, _ := stringState.CreateToken(currentIndentId, p,p+len(possibleIndent), true)
					baseToken.SetNext(newToken)
					newStringState.Advance(len(indentString))

					//we copy the indentation array
					newIndents := make([][]byte, i+1)
					copy(newIndents, possibleIndents[:i+1])
					newStringState.Indents = newIndents
					break Loop
				}
				i -= 1
			}
		}
		return newStringState, baseToken, nil
	}

	return pg.generator.parser(indentParser, "$indent", rule, false), nil
}

func (pg *StringParserPlugin) canProceed(prefixList [][]byte, state *StringState) bool {
	for i := range prefixList {
		bytePrefix := prefixList[i]
		if state.HasPrefix(bytePrefix) {
			return true
		}
	}
	return false
}

func (pg *StringParserPlugin) parser(parser Parser, name string, rule interface{}, emit bool) Parser {
	prefixes := pg.generator.getPrefixes(rule, set.HashSet{make(map[interface{}]bool)}, nil, true)
	var prefixByteList [][]byte
	var prefixStringList []string
    prefixSet := new(set.BitSet)
	ignore := false
    prefixSet.Initialize(pg.PrefixGrammar)
	usePrefixes := false
	if len(name) >= 2 && name[0:2] == "__" {
		ignore = true
	}
	if prefixes != nil {
		hashPrefixes, ok := prefixes.(*set.HashSet)
		if !ok {
			panic("Expected hash prefixes!")
		}
		if hashPrefixes.Contains(nil) || prefixes.N() == 0 || hashPrefixes.Contains(false) {
			usePrefixes = false
		} else {
			prefixesList := hashPrefixes.AsList()
			prefixStringList = make([]string, 0, len(prefixesList))
			prefixByteList = make([][]byte, 0, len(prefixesList))
			for i := range prefixesList {
				prefixString, ok := (prefixesList[i]).(string)
				if !ok {
					continue
				}
                prefixSet.Add(prefixString)
				prefixStringList = append(prefixStringList, prefixString)
				prefixByteList = append(prefixByteList, []byte(prefixString))
			}
			usePrefixes = true
		}
	}

	tokenId := pg.TokenIds.GetOrAdd(name)

	pe :=  &ParserError{fmt.Sprintf("Cannot proceed with a token %s",name)}

	wrappedParser := func(state State) (State, BaseToken, error) {

		stringState, ok := state.(*StringState)

		if !ok {
			return state, nil, &ParserError{"Invalid state: Expceted a string state in wrapper"}

		}
		level := stringState.Level

		if stringState.Debug {
			fmt.Println(strings.Repeat(" ",level),"Entering",name)
		}

        if stringState.CurrentPrefixSet == nil || stringState.CurrentPrefixSet.Pos != stringState.Pos {
            pg.updateCurrentPrefixSet(stringState)
        }

		if usePrefixes && !stringState.CurrentPrefixSet.Set.Intersects(prefixSet) {
			if stringState.Debug {

				fmt.Println(strings.Repeat(" ",level),"  >>> Cannot proceeed",string(stringState.Context.S[stringState.Pos:stringState.Pos+3]),prefixStringList)
			}
			return state, nil, pe
		}

		stringState.Level += 1

		pos := stringState.Pos
		newState, token, err := parser(state)

		stringState.Level -= 1

		if err != nil {
			if stringState.Debug {
				fmt.Println(strings.Repeat(" ",level),"Error:",name,":",err)
			}
			stringState.Context.Errors += 1
			stringState.Pos = pos
			return stringState, nil, err
		}

		if stringState.Debug {
			fmt.Println(strings.Repeat(" ",level),"Success:",name)
		}

		newStringState, ok := newState.(*StringState)
		if !ok {
			panic("Expected a string state!")
		}
		newStringState.Context.Calls += 1

		if emit {
			newToken, _ := newStringState.CreateToken(tokenId, pos, newStringState.Pos, ignore)
			if token != nil {
				stringToken, ok := token.(*Token)
				if ! ok {
					panic("Expected a string token!")
				}
				stringToken.Parent = newToken
				currentToken := stringToken
				for currentToken.Next != nil{
					currentToken = currentToken.Next
					currentToken.Parent = newToken
				}
				newToken.Children = stringToken
			}
			return newStringState, newToken, nil
		}
		return newStringState, token , nil
	}
	return wrappedParser
}

func (pg *StringParserPlugin) compileRegex(rule interface{}) (Parser, error) {

	ruleDict, ok := rule.(map[interface{}]interface{})

	if !ok {
		return nil, &ParserError{"Expected a dictionary"}
	}

	regexRule, ok := ruleDict["$regex"]

	if !ok {
		return nil, &ParserError{"regex not found"}
	}

	regexString, ok := regexRule.(string)

	if !ok {
		return nil, &ParserError{"Regex not a string"}
	}

	compiledRegex, err := regexp.Compile("(?s)^" + regexString)

	if err != nil {
		return nil, err
	}

	regexParser := func(state State) (State, BaseToken, error) {

		stringState, ok := state.(*StringState)

		if !ok {
			return state, nil, &ParserError{"Invalid state!!!!"}
		}

		result := compiledRegex.Find(stringState.Context.S[stringState.Pos:])

		if len(result) == 0 {
			return state, nil, &ParserError{"Regex did not match!"}
		}
		stringState.Advance(len(result))
		return stringState, nil, nil
	}

	return pg.generator.parser(regexParser, "regex", rule, true), nil
}

func (pg *StringParserPlugin) compileLiteral(rule interface{}) (Parser, error) {

	ruleDict, ok := rule.(map[interface{}]interface{})

	if !ok {
		return nil, &ParserError{"Expected a dictionary"}
	}

	literalRule, ok := ruleDict["$literal"]

	if !ok {
		return nil, &ParserError{"literal not found"}
	}

	literal, ok := literalRule.(string)

	replaceFunc := func (s string) string {
		return fmt.Sprintf(s)
	}

	compiledRegex, err := regexp.Compile("\\[tnsr]")

	if err != nil {
		return nil, err
	}

	literal = compiledRegex.ReplaceAllStringFunc(literal, replaceFunc)

	if !ok {
		return nil, &CompilerError{"Not a string!"}
	}

	if len(literal) == 0 {
		return nil, &CompilerError{"Literal cannot be empty!"}
	}

	if !ok {
		return nil, &ParserError{"Literal not a byte string"}
	}

	literalParser := func(state State) (State, BaseToken, error) {

		stringState, ok := state.(*StringState)

		if !ok {
			return state, nil, &ParserError{"Invalid state!!!!"}
		}

		//we don't even check for the literal as we know it will match
		if stringState.Debug {
			fmt.Println(strings.Repeat(" ",stringState.Level),"Advancing:",literal,len(literal))
		}
		stringState.Advance(len(literal))
		return stringState, nil, nil
	}

	return pg.generator.parser(literalParser, literal, rule, true), nil
}

func (pg *StringParserPlugin) compileEof(rule interface{}) (Parser, error) {

	eofParser := func(state State) (State, BaseToken, error) {

		stringState, ok := state.(*StringState)

		if !ok {
			return state, nil, &ParserError{"Invalid state at eof!!!!"}
		}

		if stringState.Pos < stringState.N {
			return stringState, nil, &ParserError{"Not at the end of input!"}
		}
		return state, nil, nil
	}

	return pg.generator.parser(eofParser, "eof", rule, true), nil
}

func (pg *StringParserPlugin) resolveRule(generator *ParserGenerator, name string) (func(rule interface{}) (Parser, error), bool) {
	switch name {
	case "$indent":
		return pg.compileIndent, true
	case "$eof":
		return pg.compileEof, true
	case "$literal":
		return pg.compileLiteral, true
	case "$regex":
		return pg.compileRegex, true
	default:
		literalParser := func(rule interface{}) (Parser, error) {
			literalRule := make(map[interface{}]interface{})
			literalRule["$literal"] = name
			return pg.compileLiteral(literalRule)
		}
		return literalParser, true
	}
	return nil, false
}
