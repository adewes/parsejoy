package main

import (
	"fmt"
	"sort"
	"regexp"
	"strings"
)

type Token[T any] struct {
	Type string
	Value T
}

type SemanticValue[T Tokenlike] struct {
	Value T
	Children []*SemanticValue[T]
}

func (s *SemanticValue[T]) PrettyString() string {
	return s.prettyString(0)
}

func (s *SemanticValue[T]) prettyString(indent int) string {
	output := ""
	if s.Value != *new(T) {
		output += fmt.Sprintf("%s%v\n", strings.Repeat(" ", indent), s.Value)
	}
	for _, child := range s.Children {
		output += child.prettyString(indent + 1)
	}
	return output
}

func (s SemanticValue[T]) String() string {

	children := make([]string, len(s.Children))

	for i, child := range s.Children {
		children[i] = child.String()
	}

	if s.Value == *new(T) {
		return fmt.Sprintf("sv(%s)", strings.Join(children, ", "))
	}

	if len(children) == 0 {
		return fmt.Sprintf("sv(%v)", s.Value)
	}

	return fmt.Sprintf("sv(%v, %s)", s.Value, strings.Join(children, ", "))
}

func (s *SemanticValue[T]) Equals(other *SemanticValue[T]) bool {

	if s.Value != other.Value {
		return false
	}

	if len(s.Children) != len(other.Children) {
		return false
	}

	for i, child := range s.Children {
		if !child.Equals(other.Children[i]) {
			return false
		}
	}

	return true
}

type Parent[T Tokenlike] struct {
	SemanticValue *SemanticValue[T]
	Head          *Head[T]
}

func (p Parent[T]) String() string {
	return fmt.Sprintf("parent(%s, %s)", p.SemanticValue.String(), p.Head.String())
}

type Head[T Tokenlike] struct {
	State    int
	Position int
	Parents  []*Parent[T]
}

func (h *Head[T]) SemanticValue() *SemanticValue[T] {
	if len(h.Parents) == 0 {
		return nil
	}
	return h.Parents[0].SemanticValue
}

func (h Head[T]) String() string {
	parents := make([]string, len(h.Parents))

	for i, parent := range h.Parents {
		parents[i] = parent.String()
	}

	return fmt.Sprintf("head(%d, %d, %s)", h.Position, h.State, strings.Join(parents, ", "))
}

func (h *Head[T]) Equals(other *Head[T]) bool {
	// to do: compare parents (?)
	return h.State == other.State && h.Position == other.Position
}

type State struct {
	State    int
	Position int
}

type Input[T Tokenlike] interface {
	Len() int
	At(pos int) (T, int)
	HasPrefix(pos int, prefix T) (bool, int)
}

type StringInput struct {
	Value string
}

func MakeStringInput(input string) *StringInput {
	return &StringInput{
		Value: input,
	}
}

func (s *StringInput) MatchRegex(re *regexp.Regexp, pos int) string {
	return re.FindString(s.Value[pos:])
}

func (s *StringInput) Len() int {
	return len(s.Value)
}

func (s *StringInput) At(pos int) (string, int) {

	if pos >= len(s.Value) {
		return "", 0
	}

	return s.Value[pos:pos+1], 1
}

func (s *StringInput) HasPrefix(pos int, prefix string) (bool, int) {
	if prefix == "" {
		return pos >= len(s.Value), 0
	}
	return strings.HasPrefix(s.Value[pos:], prefix), len(prefix)
}


type IntInput struct {
	Values []int
}

func (i *IntInput) Len() int {
	return len(i.Values)
}

func (i *IntInput) HasPrefix(pos int, prefix int) (bool, int) {
	return pos < len(i.Values) && i.Values[pos] == prefix, 1
}

type Tokenlike interface {
	comparable
	byte | int | string
}

func (s State) String() string {
	return fmt.Sprintf("(%d, %d)", s.State, s.Position)
}

type Grammar[T Tokenlike] [][]any
type Stack []*State

func (s Stack) Len() int {
	return len(s)
}

func MakeStack(states []*State) Stack {
	stack := make(Stack, 0, len(states))
outer:
	for _, state := range states {
		for _, es := range stack {
			if es.Position == state.Position && es.State == state.State {
				continue outer
			}
		}
		stack = append(stack, state)
	}
	sort.Sort(stack)
	return stack
}

func (s Stack) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Stack) Less(i, j int) bool {
	if s[i].State < s[j].State {
		return true
	}
	if s[i].State > s[j].State {
		return false
	}
	// rules are identical
	if s[i].Position < s[j].Position {
		// position is smaller
		return true
	}
	// position is larger or the same
	return false
}

func (s Stack) Add(states []*State) Stack {
	newStack := s
outer:
	for _, state := range states {
		for _, es := range newStack {
			if es.Position == state.Position && es.State == state.State {
				// we already have this
				continue outer
			}
		}
		newStack = append(newStack, state)
	}
	sort.Sort(newStack)

	return newStack
}

func (s Stack) String() string {

	states := make([]string, len(s))

	for i, state := range s {
		states[i] = state.String()
	}

	return fmt.Sprintf("{%s}", strings.Join(states, ", "))
}

type TransitionFunction[T Tokenlike] func(pos int, input Input[T]) (bool, T, int)

type Computation[T Tokenlike] struct {
	State int
	Function TransitionFunction[T]
}

type Parser[T Tokenlike] struct {
	NonTerminals map[T]bool
	NonTerminalsList []T
	Terminals    map[T]bool
	Transitions  map[int]map[T]int
	Computations map[int][]*Computation[T]
	Reductions   map[int][]int
	Stacks       []Stack
	Grammar      Grammar[T]
	StartSymbol  T
	Debug        bool
}

func MakeParser[T Tokenlike](grammar Grammar[T]) (*Parser[T], error) {
	parser := &Parser[T]{
		Grammar: grammar,
	}

	if err := parser.generateStatesAndTransitions(); err != nil {
		return nil, err
	}

	for state, _ := range parser.Transitions {
		if _, ok := parser.Reductions[state]; ok {
			// we have a shift-reduce conflict ?
			return parser, fmt.Errorf("shift-reduce conflict in state %d", state)
		}
	}

	return parser, nil
}

func (p *Parser[T]) Rules() string {

	s := "Rules:\n"

	for i, rule := range p.Grammar {

		rhs := make([]string, len(rule)-1)

		for j:=1; j < len(rule); j++ {
			rhs[j-1] = fmt.Sprintf("%v", rule[j])
		}

		s += fmt.Sprintf("%d: %v -> %s\n", i, rule[0], strings.Join(rhs, ", "))
	}

	s += "Transitions:\n"

	keys := make([]int, len(p.Transitions))
	i := 0

	for key, _ := range p.Transitions {
		keys[i] = key
		i++
	}

	sort.Ints(keys)

	for i:=0; i<len(keys); i++ {

		transitions := p.Transitions[keys[i]]
		for symbol, j := range transitions {
			s += fmt.Sprintf("%d + '%v' -> %d\n",keys[i], symbol, j)
		}
	}

	s += "Reductions:\n"

	keys = make([]int, len(p.Reductions))
	i = 0

	for key, _ := range p.Reductions {
		keys[i] = key
		i++
	}

	sort.Ints(keys)

	for i:=0;i<len(keys); i++ {
		reductions, ok := p.Reductions[keys[i]]

		if !ok {
			continue
		}

		sl := make([]string, len(reductions))

		for j, r := range reductions {
			sl[j] = fmt.Sprintf("%v(%d)", p.NonTerminalsList[r], r)
		}

		s += fmt.Sprintf("%d -> %s\n", keys[i], strings.Join(sl, ", "))
	}

	return s
} 

// Returns all rules for the given nonterminal
func (p *Parser[T]) getRulesForNonTerminal(nonTerminal T) []int {
	matchingStates := make([]int, 0)
	for i, rule := range p.Grammar {
		if rule[0] == nonTerminal {
			matchingStates = append(matchingStates, i)
		}
	}
	return matchingStates
}

// returns all rules that can follow a given rule
func (p *Parser[T]) getClosure(rule int, pos int) []*State {
	closure := make(map[int]bool)
	rulesToExamine := make([]int, 0)
	if pos >= len(p.Grammar[rule])-1 {
		// empty closure
		return []*State{}
	}
	// to do: bounds checking
	item := p.Grammar[rule][pos+1]

	switch vt := item.(type) {
	case T:
		if _, ok := p.NonTerminals[vt]; ok {
			rulesForNonTerminal := p.getRulesForNonTerminal(vt)
			for _, rule := range rulesForNonTerminal {
				closure[rule] = true
			}
			rulesToExamine = rulesForNonTerminal
		}
	}


	for {
		if len(rulesToExamine) == 0 {
			break
		}

		rule := rulesToExamine[0]
		rulesToExamine = rulesToExamine[1:]

		if len(p.Grammar[rule]) == 1 {
			// nothing to add here
			continue
		}

		// we look at the leftmost symbol of the rule
		item := p.Grammar[rule][1]

		switch vt := item.(type) {
		case T:
			if _, ok := p.NonTerminals[vt]; ok {
				newRules := p.getRulesForNonTerminal(vt)
				for _, newRule := range newRules {
					if _, ok := closure[newRule]; !ok {
						rulesToExamine = append(rulesToExamine, newRule)
						closure[newRule] = true
					}
				}
			}
		}

	}

	stackStates := make([]*State, 0)

	for rule, _ := range closure {
		stackStates = append(stackStates, &State{rule, 0})
	}

	return stackStates
}

func stackIndex(stack Stack, stacks []Stack) int {
outer:
	for i, indexState := range stacks {
		if len(stack) != len(indexState) {
			continue
		}
		for j, indexStateValue := range indexState {
			if indexStateValue.State != stack[j].State || indexStateValue.Position != stack[j].Position {
				continue outer
			}
		}
		// it's a match
		return i
	}
	return -1
}

type FunctionState[T Tokenlike] struct {
	Function TransitionFunction[T]
	State *State
}

func (p *Parser[T]) extendState(stack []*State) error {

	j := stackIndex(stack, p.Stacks)

	if j == -1 {
		return fmt.Errorf("unknown stack")
	}

	rulesBySymbol := make(map[T][]*State)
	rulesByFunction := make([]*FunctionState[T], 0)
	rulesToReduce := make([]int, 0)

	for _, state := range stack {
		if state.Position < len(p.Grammar[state.State])-1 {
			// this state is not at the end of a rule
			symbol := p.Grammar[state.State][state.Position+1]

			switch vt := symbol.(type) {
			case T:
				rulesBySymbol[vt] = append(rulesBySymbol[vt], &State{state.State, state.Position + 1})
			case TransitionFunction[T]:
				rulesByFunction = append(rulesByFunction, &FunctionState[T]{vt, &State{state.State, state.Position + 1}})
			}

		} else {
			// this state is at the end of a rule, it leads to a reduction
			rulesToReduce = append(rulesToReduce, state.State)
		}
	}


	if len(rulesToReduce) > 0 {
		p.Reductions[j] = rulesToReduce
	}

	addStates := func(states []*State) int {
		newStack := MakeStack(states)

		for _, state := range states {
			newStack = newStack.Add(p.getClosure(state.State, state.Position))
		}

		i := stackIndex(newStack, p.Stacks)

		if i == -1 {
			i = len(p.Stacks)
			p.Stacks = append(p.Stacks, newStack)
			p.extendState(newStack)
		}

		return i

	}

	for _, functionState := range rulesByFunction {

		i := addStates([]*State{functionState.State})

		if _, ok := p.Computations[j]; !ok {
			p.Computations[j] = make([]*Computation[T], 0)
		}

		p.Computations[j] = append(p.Computations[j], &Computation[T]{
			State: i,
			Function: functionState.Function,
		})

	}

	for symbol, states := range rulesBySymbol {

		i := addStates(states)

		if _, ok := p.Transitions[j]; !ok {
			p.Transitions[j] = make(map[T]int)
		}

		p.Transitions[j][symbol] = i
	}

	return nil

}

func (p *Parser[T]) generateStatesAndTransitions() error {
	p.NonTerminals = make(map[T]bool)
	p.NonTerminalsList = make([]T, len(p.Grammar))

	for i, rule := range p.Grammar {

		if item, ok := rule[0].(T); !ok {
			return fmt.Errorf("invalid grammar left-side rule!")
		} else {
			p.NonTerminals[item] = true
			p.NonTerminalsList[i] = item
			if i == 0 {
				p.StartSymbol = item
			}
		}

	}

	p.Transitions = make(map[int]map[T]int)
	p.Computations = make(map[int][]*Computation[T])
	p.Reductions = make(map[int][]int)

	initialState := Stack{{0, 0}}
	initialState = initialState.Add(p.getClosure(0, 0))
	p.Stacks = []Stack{initialState}

	return p.extendState(initialState)
}

func getPaths[T Tokenlike](head *Head[T], depth int) [][]*Parent[T] {
	if depth == 0 {
		return [][]*Parent[T]{
			[]*Parent[T]{
				&Parent[T]{
					SemanticValue: &SemanticValue[T]{
						Value: *new(T),
					},
					Head:          head,
				},
			},
		}
	}

	paths := [][]*Parent[T]{}

	for _, parent := range head.Parents {
		parentPaths := getPaths[T](parent.Head, depth-1)
		for _, parentPath := range parentPaths {
			paths = append(paths, append([]*Parent[T]{&Parent[T]{SemanticValue: parent.SemanticValue, Head: head}}, parentPath...))
		}
	}

	return paths
}

func getStackHead[T Tokenlike](heads []*Head[T], state int) *Head[T] {
	for _, head := range heads {
		if head.State == state {
			return head
		}
	}
	return nil
}

func hasStackHead[T Tokenlike](head *Head[T], heads []*Head[T]) bool {
	for _, existingHead := range heads {
		if head.State == existingHead.State && head.Position == existingHead.Position {
			// to do: check parents
			return true
		}
	}
	return false
}

func hasParent[T Tokenlike](parent *Parent[T], parents []*Parent[T]) bool {
	for _, existingParent := range parents {
		if existingParent.Head.Equals(parent.Head) {
			return true
		}
	}
	return false
}

func (p *Parser[T]) Run(input Input[T]) []*SemanticValue[T] {

	// we start in the 0 state
	states := make([]int, 1, 10000)
	symbols := make([]*SemanticValue[T], 0, 10)
	lookahead := make([]*SemanticValue[T], 0, 10)
	pos := 0

	for {

		// reduce
		reduce: 
		for {

			state := states[len(states)-1]
			reductions, ok := p.Reductions[state]

			if !ok {
				// no further reductions for this state
				break reduce
			}

			for _, reduction := range reductions {
				// number of right-side symbols in this reduction rule
				rlen := len(p.Grammar[reduction]) - 1

				if p.Debug {
					fmt.Printf("Reducing with rule %v (len: %d)\n", p.NonTerminalsList[reduction], len(p.Grammar[reduction])-1)
				}

				children := make([]*SemanticValue[T], rlen)
				copy(children, symbols[len(symbols)-rlen:])

				semanticValue := &SemanticValue[T]{
					Value: p.NonTerminalsList[reduction],
					Children: children,
				}
				// we remove the combined symbol from the list of symbols
				symbols = symbols[:len(symbols)-rlen]
				// we add the combined symbol to the lookahead value
				lookahead = append([]*SemanticValue[T]{semanticValue}, lookahead...)
				// we remove the combined states
				states = states[:len(states)-rlen]
				if p.Debug {
					fmt.Println(states, "lookahead:", lookahead, "symbols:", symbols)
				}
			}

			// we always prefer to shift
			break

		}

		// shift

		shifted := false

		shift:
		for {
			state := states[len(states)-1]
			transitions := p.Transitions[state]

			if p.Debug {
				fmt.Println("Shifting:", state, lookahead, symbols)
			}

			// we check if we have a matching symbol in the lookahead list
			if len(lookahead) > 0 {
				if ntt, ok := transitions[lookahead[0].Value];ok {
					if p.Debug {
						fmt.Printf("Shifting in symbol %v (%d)\n", lookahead[0].Value, ntt)
					}
					states = append(states, ntt)
					state = ntt
					// we add the lookahead value to the end of the symbols
					symbols = append(symbols, lookahead[0])
					// we remove the lookahead value from the lookaheads
					lookahead = lookahead[1:]
					// we mark as shifted
					shifted = true
					// we try to shift again
					continue shift
				}
			}

			// finally, we check computed transitions...
			computations := p.Computations[state]
			for _, computation := range computations {
				if ok, t, l := computation.Function(pos, input); ok {
					// we can shift in a computation

					if p.Debug {
						fmt.Printf("Shifting in computed value %v (%d)\n", t, computation.State)
					}
	
					states = append(states, computation.State)
	
					// we advance the position
					pos += l

					symbols = append(symbols, &SemanticValue[T]{
						Value: t,
					})

					shifted = true

					continue shift
				}
			}

			// we check if one of the transition matches with the current input

			for t, tr := range transitions {
				if ok, l := input.HasPrefix(pos, t); ok {
					if p.Debug {
						fmt.Printf("Shifting in terminal %v (%d)\n", t, tr)
					}
					// we create a new semantic value and add it to the symbols
					symbols = append(symbols, &SemanticValue[T]{
						Value: t,
					})
					// we advance the position
					pos += l
					// we update the states list
					states = append(states, tr)
					// we mark as shifted
					shifted = true
					// we try to shift again
					continue shift
				}
			}

			// we can't shift anymore
			break shift
		}

		if pos > input.Len() || !shifted {
			break
		}

	}

	return lookahead
}

var termGrammar = Grammar[string]{
	{"S", "term", ""},
	{"term", "factor"},
	{"term", "factor", "+", "term"},
	{"factor", "s", "times", "factor"},
	{"factor", "s"},
	{"s", "number"},
	{"s", "symbol"},
	{"symbol", "a"},
	{"times", "*"},
	{"number", "digits"},
	{"digits", "digits", "digit"},
	{"digits", "digit"},
	{"digit", "1"},
	{"digit", "2"},
	{"digit", "3"},
	{"digit", "0"},
}

var eGrammar = Grammar[string]{
    {"S","e",""},
    {"e","e","+","e"},
    {"e","b"},
}

func CompileRegex[T Tokenlike](value string) TransitionFunction[T] {

	re, err := regexp.Compile(value)

	if err != nil {
		panic(err)
	}

	return func(pos int, input Input[T]) (bool, T, int) {
		switch vt := input.(type) {
		case Input[string]:
			if result := vt.(*StringInput).MatchRegex(re, pos); result != "" {
				return true, any(result).(T), len(result)
			}
		default:
			panic("regexp undefined for this type")
		}
		return false, *new(T), 0
	}
}

var grammarGrammar = Grammar[string]{
	{"S", "ows", "[]rules", "ows", ""},
	{"[]rules"},
	{"[]rules", "[]rules", "{}rule"},
	{"{}rule", "ows", ".name", "[]args", "ows", "->", "patterns", ";"},
	{"[]args"},
	{"[]args", "(", "arglist", ")"},
	{"arglist"},
	{"arglist", "arglist", ",", "|arg"},
	{"arglist", "|arg"},
	{"|arg", CompileRegex[string]("[a-z]+")},
	{"patterns", "alternatives"},
	{"patterns", "[]patternlist"},
	{"alternatives", "[]alternativelist"},
	{"[]alternativelist", "[]patternlist"},
	{"[]alternativelist", "[]alternativelist", "|", "[]patternlist"},
	{"[]patternlist", "[]patternlist", ",", "{}pattern"},
	{"[]patternlist", "{}pattern"},
	{"[]patternlist", "ows"},
	{"{}pattern", "ows", "pattern-type", "ows"},
	{"pattern-type", ".name"},
	{"pattern-type", "expression"},
	{"pattern-type", ".reference"},
	{"pattern-type", ".literal"},
	{"pattern-type", ".regex"},
	{"pattern-type", ":end"},
	{"expression", "(", "ows", "expression-value", "ows", ")"},
	{"expression-value", "[]expr-alternatives"},
	{"[]expr-alternatives", "[]expr-alternatives", "ows", "|", "ows", "{}expr-alternative"},
	{"[]expr-alternatives", "{}expr-alternative"},
	{"{}expr-alternative", ".name"},
	{".reference", "\\", ":reference-value"},
	{":reference-value", CompileRegex[string]("[0-9]+")},
	{":end", "$"},
	{".name", "|name-value"},
	{"|name-value", CompileRegex[string](`(:|\[\]|\{\}|\.|\|)?[^\#\s\|\[\]\|\.\:\;\,\"\'\)\(\\\-]+`)},
	{".literal", "\"", ":literal-value", "\"", ":literal-suffix"},
	{":literal-value", CompileRegex[string](`(\\.|[^\"])*`)},
	{".regex", "re:", "|regex-value"},
	{"|regex-value", CompileRegex[string](`(\\.|[^\;\,\n])*`)},
	{":literal-suffix"},
	{":literal-suffix", "_foo"},
	{"optional_newline"},
	{"optional_newline", "newline"},
	{"newline", "\n"},
	{"newline-or-end", "newline"},
	{"newline-or-end", ""},
	{"ows"},
	{"ows", "ws"},
	{"ws", "ws", "wsc"},
	{"ws", "wsc"},
	{"wsc", "comment"},
	{"wsc", " "},
	{"wsc", "\t"},
	{"wsc", "\n"},
	{"comment", "#", "anything", "newline-or-end"},
	{"anything", CompileRegex[string](`[^\n]*`)},
}



func main() {

	parser, err := MakeParser[string](termGrammar)

	fmt.Println(parser.Rules())

	if err != nil {
		fmt.Printf("cannot build parser: %v\n", err)
		return
	}

	parser.Debug = true


	for _, stack := range parser.Stacks {
		fmt.Println(stack)
	}

	input := MakeStringInput("a*10+22+a*a*a*a+a")

	semanticValues := parser.Run(input)

	fmt.Printf("Got %d semantic values\n", len(semanticValues))
	fmt.Println(semanticValues[0].PrettyString())
	//fmt.Println(acceptedHeads[0].SemanticValue().PrettyString())


	parser, err = MakeParser[string](eGrammar)

	if err != nil {
		fmt.Println(err)
		return
	}

	eStr := "b"

	for i :=0;i < 1; i++ {
		eStr += "+b"
	}

	fmt.Println("ready")

	fmt.Println(parser.Rules())

	input = MakeStringInput(eStr)
	semanticValues = parser.Run(input)
	fmt.Printf("Got %d semantic values (eGrammar)\n", len(semanticValues))
	fmt.Println(semanticValues[0].PrettyString())
	// fmt.Println(acceptedHeads[0].SemanticValue().PrettyString())

	grammarParser, err := MakeParser[string](grammarGrammar)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("grammar parser is ready")
	}

	fmt.Println(grammarParser.Rules())

	grammarStr := ""

	for i:=0; i<1; i++ {
		grammarStr += "baz -> bar, bar, bar, bar; bar -> bam;"
	}

	grammarProgram := MakeStringInput(grammarStr)

	grammarParser.Debug = true

	fmt.Println("Running...")

	semanticValues = grammarParser.Run(grammarProgram)
	fmt.Printf("Got %d semantic values\n", len(semanticValues))

	for _, semanticValue := range semanticValues {
		fmt.Println(semanticValue.PrettyString())	
	}

}
