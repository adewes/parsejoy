package main

import (
	"fmt"
	"sort"
	"strings"
)

type Parent[T Tokenlike] struct {
	SemanticValue T
	Head          *Head[T]
}

type Head[T Tokenlike] struct {
	State    int
	Position int
	Parents  []*Parent[T]
}

func (h *Head[T]) Equals(other *Head[T]) bool {
	// to do: compare parents (?)
	return h.State == other.State && h.Position == other.Position
}

type State struct {
	State    int
	Position int
}

type Input[T Tokenlike] struct {
	input []T
}

func MakeInput[T Tokenlike](v []T) *Input[T] {
	return &Input[T]{
		input: v,
	}
}

func (i *Input[T]) Len() int {
	return len(i.input)
}

func (i *Input[T]) At(pos int) []T {
	return i.input[pos : pos+1]
}

func (i *Input[T]) From(pos int) []T {
	return i.input[pos:]
}

func (i *Input[T]) HasPrefix(pos int, prefix T) bool {
	switch vt := any(i.input).(type) {
	case []string:
		return vt[pos] == any(prefix).(string)
	case []int:
		pv := any(prefix).([]int)
		for i, v := range vt[pos:] {
			if i >= len(pv) {
				return true
			}
			if v != pv[i] {
				return false
			}
		}
	}

	return false
}

type Tokenlike interface {
	comparable
	byte | int | string
}

func (s State) String() string {
	return fmt.Sprintf("(%d, %d)", s.State, s.Position)
}

type Grammar[T Tokenlike] [][]T
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

type Parser[T Tokenlike] struct {
	NonTerminals map[T]bool
	Transitions  map[int]map[T]int
	Reductions   map[int][]int
	Stacks       []Stack
	Grammar      Grammar[T]
	Debug        bool
}

func MakeParser[T Tokenlike](grammar Grammar[T]) (*Parser[T], error) {
	parser := &Parser[T]{
		Grammar: grammar,
	}

	if err := parser.generateStatesAndTransitions(); err != nil {
		return nil, err
	}

	return parser, nil
}

func (p *Parser[T]) getStatesForNonTerminal(nonTerminal T) []int {
	matchingStates := make([]int, 0)
	for i, rule := range p.Grammar {
		if rule[0] == nonTerminal {
			matchingStates = append(matchingStates, i)
		}
	}
	return matchingStates
}

func (p *Parser[T]) getClosure(rule int, pos int) []*State {
	closure := make(map[int]bool)
	rulesToExamine := make([]int, 0)
	if pos >= len(p.Grammar[rule])-1 {
		// empty closure
		return []*State{}
	}
	// to do: bounds checking
	item := p.Grammar[rule][pos+1]

	if _, ok := p.NonTerminals[item]; ok {
		rulesForNonTerminal := p.getStatesForNonTerminal(item)
		for _, rule := range rulesForNonTerminal {
			closure[rule] = true
		}
		rulesToExamine = rulesForNonTerminal
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

		// to do: bounds checking
		if _, ok := p.NonTerminals[p.Grammar[rule][1]]; ok {
			newStates := p.getStatesForNonTerminal(p.Grammar[rule][1])
			for _, newState := range newStates {
				if _, ok := closure[newState]; !ok {
					rulesToExamine = append(rulesToExamine, newState)
					closure[newState] = true
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

func (p *Parser[T]) extendState(stack []*State) error {

	j := stackIndex(stack, p.Stacks)

	if j == -1 {
		return fmt.Errorf("unknown stack")
	}

	rulesBySymbol := make(map[T][]*State)
	rulesToReduce := make([]int, 0)

	for _, state := range stack {
		if len(p.Grammar[state.State]) > state.Position+1 {
			symbol := p.Grammar[state.State][state.Position+1]
			rulesBySymbol[symbol] = append(rulesBySymbol[symbol], &State{state.State, state.Position + 1})
		} else {
			rulesToReduce = append(rulesToReduce, state.State)
		}
	}


	if len(rulesToReduce) > 0 {
		p.Reductions[j] = rulesToReduce
	}

	for symbol, states := range rulesBySymbol {
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

		if _, ok := p.Transitions[j]; !ok {
			p.Transitions[j] = make(map[T]int)
		}

		p.Transitions[j][symbol] = i
	}

	return nil

}

func (p *Parser[T]) generateStatesAndTransitions() error {
	p.NonTerminals = make(map[T]bool)

	for _, rule := range p.Grammar {
		p.NonTerminals[rule[0]] = true
	}

	p.Transitions = make(map[int]map[T]int)
	p.Reductions = make(map[int][]int)
	p.Stacks = make([]Stack, 0)

	initialState := Stack{{0, 0}}
	initialState = initialState.Add(p.getClosure(0, 0))
	p.Stacks = append(p.Stacks, initialState)

	return p.extendState(initialState)
}

func (p *Parser[T]) getPaths(head *Head[T], depth int) [][]*Parent[T] {
	if depth == 0 {
		return [][]*Parent[T]{
			[]*Parent[T]{
				&Parent[T]{
					SemanticValue: *new(T),
					Head:          head,
				},
			},
		}
	}

	paths := [][]*Parent[T]{}

	for _, parent := range head.Parents {
		parentPaths := p.getPaths(parent.Head, depth-1)
		for _, parentPath := range parentPaths {
			paths = append(paths, append([]*Parent[T]{&Parent[T]{SemanticValue: parent.SemanticValue, Head: head}}, parentPath...))
		}
	}

	return paths
}

func (p *Parser[T]) getStackHead(heads []*Head[T], state int) *Head[T] {
	for _, head := range heads {
		if head.State == state {
			return head
		}
	}
	return nil
}

func (p *Parser[T]) shiftStackHeads(stackHeads []*Head[T], input *Input[T]) []*Head[T] {
	newStackHeads := make([]*Head[T], 0)

	for {

		fmt.Println("--.-")
		if len(stackHeads) == 0 {
			break
		}

		head := stackHeads[0]
		stackHeads = stackHeads[1:]

		var transition int
		var semanticValue T

		for value, tr := range p.Transitions[head.State] {
			if input.HasPrefix(head.Position, value) {
				semanticValue = value
				transition = tr
				break
			}
		}

		if semanticValue != *new(T) {

			parent := &Parent[T]{SemanticValue: semanticValue, Head: head}

			newStackHead := &Head[T]{
				State:    transition,
				Position: head.Position + 1,
				Parents:  []*Parent[T]{parent},
			}
			existingStackHead := p.getStackHead(newStackHeads, transition)

			if existingStackHead != nil {
				existingStackHead.Parents = append(existingStackHead.Parents, parent)
			} else {
				newStackHeads = append(newStackHeads, newStackHead)
			}
		}
	}
	return newStackHeads
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
		if existingParent.SemanticValue == parent.SemanticValue && existingParent.Head.Equals(parent.Head) {
			return true
		}
	}
	return false
}

func (p *Parser[T]) reduceStackHeads(stackHeads []*Head[T], input *Input[T]) []*Head[T] {
	newStackHeads := []*Head[T]{}
	for {
		if len(stackHeads) == 0 {
			break
		}

		fmt.Println(".---", len(stackHeads))

		head := stackHeads[0]
		stackHeads = stackHeads[1:]

		if head.Position < input.Len() && !hasStackHead(head, newStackHeads) {
			newStackHeads = append(newStackHeads, head)
		}

		if rr, ok := p.Reductions[head.State]; ok {
			for _, r := range rr {
				nonTerminal := p.Grammar[r][0]
				reduceLength := len(p.Grammar[r])-1
				paths := p.getPaths(head, reduceLength)
				for _, path := range paths {
					parent := path[len(path)-1]

					var newState int

					// we check if this is the start symbol
					if nonTerminal == p.Grammar[0][0] {
						newState = -1
					} else {
						newState = p.Transitions[parent.Head.State][nonTerminal]
					}
					existingStackHead := p.getStackHead(newStackHeads, newState)

					if existingStackHead != nil {
						if hasParent(parent, existingStackHead.Parents) {
							// to do: conflict handling
						} else {
							fmt.Println("no parent...", nonTerminal, parent.Head.Position, parent.Head.State, existingStackHead.Position, existingStackHead.State, len(existingStackHead.Parents))
							existingStackHead.Parents = append(existingStackHead.Parents, &Parent[T]{
								SemanticValue: nonTerminal,
								Head: parent.Head,
							})
							// we process the stack head again to see if we can further reduce it
							stackHeads = append([]*Head[T]{head}, stackHeads...)							
						}
					} else {
						newStackHead := &Head[T]{
							State: newState,
							Position: head.Position,
							Parents: []*Parent[T]{
								&Parent[T]{
									SemanticValue: nonTerminal,
									Head: parent.Head,
								},
							},
						}
						// we append the stack head to the new stack heads
						newStackHeads = append(newStackHeads, newStackHead)
						// we reprent the new stack head to the list of stack heads to process
						stackHeads = append([]*Head[T]{newStackHead}, stackHeads...)
					}
				}
			}
		}


	}
	return newStackHeads
}

func (p *Parser[T]) Run(input *Input[T]) []*Head[T] {

	stackHeads := []*Head[T]{
		&Head[T]{
			Position: 0,
			State: 0,
			Parents: nil,
		},
	}

	acceptedStacks := []*Head[T]{}

	for {
		newStackHeads := p.reduceStackHeads(stackHeads, input)

		for _, head := range newStackHeads {
			if head.Position == input.Len() && head.State == -1 {
				acceptedStacks = append(acceptedStacks, head)
			}
		}

		fmt.Println(len(newStackHeads),"...")

		stackHeads = p.shiftStackHeads(newStackHeads, input)

		if len(stackHeads) == 0 {
			break
		}
	}

	return acceptedStacks
}

var termGrammar = Grammar[string]{
	{"S", "term", "$"},
	{"term", "factor"},
	{"term", "factor", "+", "term"},
	{"factor", "s", "times", "factor"},
	{"factor", "s"},
	{"s", "number"},
	{"s", "symbol"},
	{"number", "digits"},
	{"digits", "digits", "digit"},
	{"digits", "digit"},
	{"digit", "1"},
	{"digit", "2"},
	{"digit", "3"},
}

var inputString = []string{"symbol", "times", "symbol", "times", "number", "+", "symbol"}

func main() {
	parser, err := MakeParser[string](termGrammar)

	if err != nil {
		fmt.Printf("cannot build parser: %v\n", err)
	} else {
		fmt.Println("it worked")
		fmt.Println(parser.Transitions)
		for _, stack := range parser.Stacks {
			fmt.Println(stack)
		}

		input := MakeInput[string](inputString)

		acceptedStacks := parser.Run(input)

		fmt.Printf("Got %d accepted stacks\n", len(acceptedStacks))
	}
}
