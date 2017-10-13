package parsejoy
import "crypto/sha1"
import "encoding/hex"
import "fmt"
import "strings"
import "7scientists.com/parsejoy/set"

type State interface{
    Copy() State
    //Pushes a new AST node onto the AST stack
    PushASTNode(node *ASTNode, state State) error
    //Pushes an AST property onto the AST stack
    PushASTProperty(property *ASTProperty, state State) error
}

/*
$ast-node: Calls PushASTNode with a new node
$ast-property: Calls PushASTProperty with a new property

PushASTNode should check which values are currently on the stack.
It should go through each of these values and make sure that it's
and ASTProperty. If it is, it should get incorporated into the
current ASTNode.

PushASTProperty should just push the property onto the current
stack.
*/

type ASTNode struct{
    //represents an AST node
    Type string
    Attributes map[string]interface{}
}

func (node *ASTNode) Initialize(nodeType string, properties map[interface{}]interface{}) {
    node.Type = nodeType
    node.Attributes = make(map[string]interface{})
}

type ASTProperty struct{
    //represents an AST property
    Name string
    List bool
    Value interface{}
}

type BaseToken interface{
    SetNext(token BaseToken)
    GetValue(state State) string
}

type CompilerError struct{
    msg string
}

type ParserError struct{
    msg string
}

func (e *CompilerError) Error() string {return e.msg}
func (e *ParserError) Error() string {return e.msg}

type Parser func(State) (State, BaseToken, error)
type ParserMap map[string]Parser

type ParserPlugin interface {
    setPrefixes(name string, prefix interface{})
    PrepareGrammar(grammar map[string]interface{})
    //getPrefixSet(state State) set.Set
    resolveRule(pg *ParserGenerator, name string) ((func (rule interface{}) (Parser, error) ) , bool)
    parser(parser Parser,name string,rule interface{},emit bool) Parser
    getPrefix(rule interface{},args interface{}) set.Set
    getEmptyPrefix(args interface{}) set.Set
    SetGenerator(generator *ParserGenerator)
}

type ParserGenerator struct {
    grammar map[string]interface{}
    plugin ParserPlugin
    parsers ParserMap
}

type ParserPluginMixin struct {
    RulePrefixes map[string]interface{}
    generator *ParserGenerator
    fingerprintIds map[string]uint
    fingerprintId uint
}

type OutcomeMixin struct {
    Outcome [][]Outcome
    OutcomeMap []uint
}

func PrintASTNode(node *ASTNode, level int) int {
	nodes := 1
	fmt.Println(strings.Repeat(" ", level*2), node.Type,":")
    for key, value := range node.Attributes {
        astNodeValue, ok := value.(*ASTNode)
        if ok {
            nodes += PrintASTTree(astNodeValue, level*2)
            continue
        }
        astNodeList, ok := value.([]interface{})
        if ok {
            fmt.Println(strings.Repeat(" ", level*2+1), key,":")
            for i, item := range astNodeList {
                astNodeItem, ok := item.(*ASTNode)
                if ok {
                    nodes += PrintASTTree(astNodeItem, level+1)
                } else {
                    fmt.Println(strings.Repeat(" ", level*2+1), key, "(", i, ")=", astNodeItem)
                }
            }
            continue
        }
        fmt.Println(strings.Repeat(" ", level*2+1), key, "=", value)
    }
	return nodes
}

func PrintASTTree(node interface{}, level int) int {
    astNode, ok := node.(*ASTNode)
    if ok {
        return PrintASTNode(astNode, level)
    }
    nodes := 1
    astNodeList, ok := node.([]interface{})
    if ok {
        for _, item := range astNodeList {
            fmt.Println(item)
            nodes += PrintASTTree(item, level)
        }
    }
    return nodes
}


func (om *OutcomeMixin) Initialize(N uint,M uint) {
    om.OutcomeMap = make([]uint,N)
    om.Outcome = make([][]Outcome,0,M)
}

func (pg *ParserPluginMixin) setPrefixes(name string, prefix interface{}) {
    pg.RulePrefixes[name] = prefix
}

func (pg *ParserPluginMixin) Initialize() {
    pg.fingerprintIds = make(map[string]uint)
    pg.RulePrefixes = make(map[string]interface{})
    pg.fingerprintId = 1
}

func (pg *ParserPluginMixin) SetGenerator (generator *ParserGenerator){
    pg.generator = generator
}

func (pg *ParserGenerator) SetPlugin(plugin ParserPlugin) {
    pg.plugin = plugin
    plugin.SetGenerator(pg)
}

func (pg *ParserGenerator) GetFingerprint (rule interface{}) string {

    data := "foobar"
    stringRule, ok := rule.(string)
    if ok {
        data = data+stringRule
    } else {
        listRule, ok := rule.([]interface{})
        if ok {
            for i := range listRule {
                element := listRule[i]
                data = data+pg.GetFingerprint(element)
            }
        } else {
            mapRule, ok := rule.(map[interface{}]interface{})

            if ok {
                for key, value := range mapRule {
                    data = data+pg.GetFingerprint(key)
                    data = data+pg.GetFingerprint(value)
                }
            }
        }
    }

    byteData := []byte(data)
    s1 := sha1.Sum(byteData)
    return hex.EncodeToString(s1[:])
}

func (pg *ParserGenerator) getPrefixes (rule interface{}, visitedRules set.HashSet, args interface{}, onlyStart bool) set.Set {

    stringRule, ok := rule.(string)

    if ok {
        if visitedRules.Contains(stringRule) {
            return pg.plugin.getEmptyPrefix(args)
        }
        visitedRules.Add(stringRule)
        grammarRule, ok := pg.grammar[stringRule]
        if ok {
            prefixes := pg.getPrefixes(grammarRule,visitedRules,args, onlyStart)
            pg.plugin.setPrefixes(stringRule, prefixes)
            return prefixes
        }
        return pg.plugin.getPrefix(stringRule,args)
    }

    listRule, ok := rule.([]interface{})

    if ok && len(listRule) > 0{
        prefixes := pg.plugin.getEmptyPrefix(args)
        for i:= range listRule  {
            ruleItem := listRule[i]
            subrulePrefixes := pg.getPrefixes(ruleItem, visitedRules,args,onlyStart)
            //if the current prefixes contain a false value (i.e. an optional token), we first remove it
            //before adding the new prefixes
            if prefixes.Contains(false) {
                _ = prefixes.Remove(false)
            }
            prefixes, _ = prefixes.Union(subrulePrefixes)
            //if we match only the start and this contains no false value, we break from the loop
            if onlyStart && !prefixes.Contains(false){
                break
            }
        }
        return prefixes
    }

    dictRule, ok := rule.(map[interface{}]interface{})

    if ok {
        if len(dictRule) == 1 {
            var key, value interface{}
            for key, value = range dictRule {
                break
            }
            stringKey, ok := key.(string)
            if !ok {
                panic("This should not happen!")
            }
            switch stringKey {
                case "$or":
                    listValue, ok := value.([]interface{})
                    if !ok || len(listValue) == 0 {
                        panic("This should not happen!")
                    }
                    prefixes := pg.plugin.getEmptyPrefix(args)
                    for i := range listValue {
                        subrule := listValue[i]
                        newPrefixes := pg.getPrefixes(subrule, visitedRules, args, onlyStart)
                        prefixes, _ = prefixes.Union(newPrefixes)
                    }
                    return prefixes
                case "$ast-prop": fallthrough
                case "$ast-node":
                    dictValue, ok := value.(map[interface{}]interface{})
                    if !ok {
                    }
                    return pg.getPrefixes(dictValue["value"], visitedRules, args, onlyStart)
                case "$and": fallthrough
                case "$repeat":
                    return pg.getPrefixes(value, visitedRules, args, onlyStart)
                case "$not":
                    nilPrefix := pg.plugin.getEmptyPrefix(args)
                    nilPrefix.Add(nil)
                    return nilPrefix
                case "$optional":
                    newPrefixes := pg.getPrefixes(value, visitedRules,args, onlyStart)
                    newPrefixes.Add(false)
                    return newPrefixes
                default:
                    return pg.plugin.getPrefix(rule,args)
            }
        }
    }
    panic("this should never happen")
}

func (pg *ParserGenerator) compileSequence (rules []interface{}) (Parser, error) {

    parsers := make([]Parser,len(rules))

    for i, rule := range rules {
        parser, ok := pg.compileRule(rule)
        if (ok != nil){
            return nil, ok
        }
        parsers[i] = parser
    }

    sequenceParser := func(state State) (State, BaseToken, error) {
        currentState := state
        var err error
        var baseToken BaseToken
        var currentToken BaseToken
        for i := range parsers{
            var newToken BaseToken
            currentState, newToken, err = parsers[i](currentState)
            if newToken != nil {
                if baseToken == nil {
                    baseToken = newToken
                }
                if currentToken != nil {
                    currentToken.SetNext(newToken)
                }
                currentToken = newToken
            }
            if err != nil{
                if baseToken != nil{
                    baseToken.SetNext(nil)
                }
                return state, nil, err
            }
        }
        return currentState, baseToken, nil
    }

    return pg.parser(sequenceParser, "seq", rules, false), nil
}

func (pg *ParserGenerator) parser(fn Parser, name string, rule interface{}, emit bool) Parser {

    if pg.plugin != nil {
        return pg.plugin.parser(fn,name,rule,emit)
    }
    wrappedParser := func(state State) (State, BaseToken, error){
        return fn(state)
    }
    return wrappedParser
}

func (pg *ParserGenerator) extractSubrule(rule interface{}, name string) (interface{}, bool) {
    ruleDict, ok := rule.(map[interface{}]interface{})

    if !ok {
        return nil, false
    }

    resolvedRule, ok := ruleDict[name]

    return resolvedRule, ok

}

func (pg *ParserGenerator) compileRepeat (rule interface{}) (Parser, error) {

    repeatRule, ok := pg.extractSubrule(rule, "$repeat")

    if !ok{
        return nil, &CompilerError{"repeat rule not found!"}
    }

    parser, err := pg.compileRule(repeatRule)

    if err != nil{
        return nil, err
    }

    pe := &ParserError{"repeat did not match once!"}

    repeatParser := func(state State) (State, BaseToken, error) {
        cnt := 0
        currentState := state
        var err error
        var baseToken BaseToken
        var currentToken BaseToken
        for {
            var newToken BaseToken
            currentState, newToken, err = parser(currentState)
            if newToken != nil {
                if baseToken == nil {
                    baseToken = newToken
                }
                if currentToken != nil {
                    currentToken.SetNext(newToken)
                }
                currentToken = newToken
            }
            if err != nil {
                break
            }
            cnt += 1
        }
        if cnt == 0 {
            if baseToken != nil{
                baseToken.SetNext(nil)
            }
            return state, nil, pe
        }
        return currentState, baseToken, nil
    }
    return pg.parser(repeatParser, "$repeat", rule, false), nil
}


func (pg *ParserGenerator) compileOr (rule interface{}) (Parser, error) {

    orRule, ok := pg.extractSubrule(rule, "$or")

    ruleArray, ok := orRule.([]interface{})

    if !ok{
        return nil, &CompilerError{"Expected a list of rules!"}
    }

    alternatives := make([]Parser,len(ruleArray))


    for i := range ruleArray {
        alternativeRule := ruleArray[i]
        compiledRule, err := pg.compileRule(alternativeRule)
        if err != nil{
            return nil, err
        }
        alternatives[i] = compiledRule
    }

    pe := &ParserError{"No alternative matched!"}


    orParser := func(state State) (State, BaseToken, error) {
        for i := range alternatives {
            newState, baseToken, err := alternatives[i](state)
            if err == nil{
                return newState, baseToken, nil
            }
        }
        return state, nil, pe
    }
    return pg.parser(orParser, "or", rule, false), nil

}


func (pg *ParserGenerator) compileAnd (rule interface{}) (Parser, error) {

    andRule, ok := pg.extractSubrule(rule, "$and")

    if !ok{
        return nil, &CompilerError{"and rule not found!"}
    }

    parser, err := pg.compileRule(andRule)

    if err != nil{
        return nil, err
    }

    pe := &ParserError{"And condition did not match!"}

    andParser := func(state State) (State, BaseToken, error) {
        stateCopy := state.Copy()
        _, _, err := parser(stateCopy)
        if err == nil {
            return state, nil, nil
        }
        return state, nil, pe
    }
    return pg.parser(andParser, "and", rule, false), nil

}

func (pg *ParserGenerator) compileNot (rule interface{}) (Parser, error) {

    notRule, ok := pg.extractSubrule(rule, "$not")

    if !ok{
        return nil, &CompilerError{"not rule not found!"}
    }

    parser, err := pg.compileRule(notRule)

    if err != nil{
        return nil, err
    }

    notParser := func(state State) (State, BaseToken, error) {
        stateCopy := state.Copy()
        _, _, err := parser(stateCopy)
        if err != nil {
            return state, nil, nil
        }
        return state, nil, &ParserError{"Not condition did match!"}
    }
    return pg.parser(notParser, "not", rule, false), nil

}

func (pg *ParserGenerator) compileOptional (rule interface{}) (Parser, error) {

    optionalRule, ok := pg.extractSubrule(rule, "$optional")

    if !ok{
        return nil, &CompilerError{"optional rule not found!"}
    }

    parser, err := pg.compileRule(optionalRule)

    if err != nil{
        return nil, err
    }

    optionalParser := func(state State) (State, BaseToken, error) {
        newState, baseToken, err := parser(state)
        if err != nil {
            return state, nil, nil
        }
        return newState, baseToken, nil
    }
    return pg.parser(optionalParser, "optional", rule, false), nil

}

func (pg *ParserGenerator) compileASTProperty(rule interface{}) (Parser, error) {

  astRule, ok := pg.extractSubrule(rule, "$ast-prop")

  if !ok{
      return nil, &CompilerError{"AST property rule not found!"}
  }

  astStringRule, ok := astRule.(map[interface{}]interface{})

  if !ok {
    return nil, &CompilerError{"Expected a map"}
  }

  valueRule, ok := astStringRule["value"]

  if !ok {
    return nil, &CompilerError{"You need to specify a value!"}
  }

  name, ok := astStringRule["name"]

  if !ok {
      return nil, &CompilerError{"You need to specify a name!"}
  }

  stringName, ok := name.(string)

  if !ok {
      return nil, &CompilerError{"Name must be a string!"}
  }

  //properties is an optional field
  asList, ok := astStringRule["as-list"]

  var asListBool bool

  if ok {
      asListBool, ok = asList.(bool)
      if !ok {
          return nil, &CompilerError{"as-list must be a boolean!"}
      }
  }

  asLiteral, ok := astStringRule["as-literal"]

  var asLiteralBool bool

  if ok {
      asLiteralBool, ok = asLiteral.(bool)
      if !ok {
          return nil, &CompilerError{"as-literal must be boolean!"}
      }
  }

  parser, err := pg.compileRule(valueRule)

  if err != nil{
      return nil, err
  }

  /*
  * Get the length of the current AST stack
  * When the parser is finished, move all new objects in the
    current AST stack into the AST property and then
    push the property itself onto the current stack.
  */
  astParser := func(state State) (State, BaseToken, error) {
      newState, baseToken, err := parser(state)
      if err == nil {
          astProperty := new(ASTProperty)
          astProperty.Name = stringName
          astProperty.List = asListBool
          if asLiteralBool {
              astProperty.Value = baseToken.GetValue(newState)
          }
          newState.PushASTProperty(astProperty, state)
      }
      return newState, baseToken, err
  }

  return astParser,nil

}

func (pg *ParserGenerator) compileASTNode(rule interface{}) (Parser, error) {

  astRule, ok := pg.extractSubrule(rule, "$ast-node")

  if !ok{
      return nil, &CompilerError{"AST-node rule not found!"}
  }

  astStringRule, ok := astRule.(map[interface{}]interface{})

  if !ok {
    return nil, &CompilerError{"Expected a map"}
  }

  valueRule, ok := astStringRule["value"]

  if !ok {
    return nil, &CompilerError{"You need to specify a value!"}
  }

  nodeType, ok := astStringRule["type"]

  if !ok {
      return nil, &CompilerError{"You need to specify an node type!"}
  }

  stringNodeType, ok := nodeType.(string)

  if !ok {
      return nil, &CompilerError{"Node type must be a string!"}
  }

  //properties is an optional field
  properties, ok := astStringRule["properties"]

  var mapProperties map[interface{}]interface{}

  if ok {
      mapProperties, ok = properties.(map[interface{}]interface{})
      if !ok {
          return nil, &CompilerError{"Properties must be a map!"}
      }
  }

  parser, err := pg.compileRule(valueRule)

  if err != nil{
      return nil, err
  }

  astParser := func(state State) (State, BaseToken, error) {
      newState, baseToken, err := parser(state)
      if err == nil {
          astNode := new(ASTNode)
          astNode.Initialize(stringNodeType, mapProperties)
          newState.PushASTNode(astNode, state)
      }
      return newState, baseToken, err
  }

  return astParser,nil

}

func (pg *ParserGenerator) resolveRule(name string) ((func (rule interface{}) (Parser, error) ) , bool) {
    switch name {
    case "$ast-node":
        return pg.compileASTNode, true
    case "$ast-prop":
        return pg.compileASTProperty, true
    case "$repeat":
        return pg.compileRepeat, true
    case "$or":
        return pg.compileOr, true
    case "$and":
        return pg.compileAnd, true
    case "$not":
        return pg.compileNot, true
    case "$optional":
        return pg.compileOptional, true
    default:
        if pg.plugin != nil {
            return pg.plugin.resolveRule(pg, name)
        }
    }
    return nil, false
}

func (pg *ParserGenerator) compileRule (rule interface{}) (Parser, error) {
    name := ""
    if rn, ok := rule.(string);ok {
        name = rn
        if parser, okr := pg.parsers[name]; okr {
            return parser, nil
        } else {
            if grammarRule, okr := pg.grammar[name]; okr {
                rule = grammarRule
            } else {

                if generator, ok := pg.resolveRule(name); ok {
                    return generator(rule)
                }
                return nil, &CompilerError{"Unknown rule: "+rn}
            }
        }
    }

    parseSubrule := func(rule map[interface{}]interface{}, name string) (Parser, error) {
        if len(rule) == 1 {
            var ruleName string
            var ok bool
            for key, _ := range rule{
                ruleName, ok = key.(string)
                break
            }
            if !ok{
                return nil,&CompilerError{"Key is not a string!"}
            }
            if generator, ok := pg.resolveRule(ruleName); ok {
                parser, err := generator(rule)
                if err != nil{
                    return nil, err
                }
                if name != "" {
                    pg.parsers[name] = pg.parser(parser, name, rule, true)
                    return pg.parsers[name], nil
                }
                return parser, nil
            } else{
                return nil, &CompilerError{"Unknown rule!"}
            }
        }
        return nil, nil
    }

    if name != "" {
        subruleParser := func(state State) (State, BaseToken, error) {
            return pg.parsers[name](state)
        }
        pg.parsers[name] = subruleParser
    }
    switch rt := rule.(type) {
    case map[interface{}]interface{}:
        if len(rt) == 1 {
            return parseSubrule(rt, name)
        }

    case []interface{}:
        if sequenceParser, err := pg.compileSequence(rt); err != nil{
            return nil, err
        } else {
            if name != "" {
                pg.parsers[name] = pg.parser(sequenceParser, name, rule, true)
                return pg.parsers[name], nil
            }
            return sequenceParser, nil
        }
    case string:
        if subrule, err := pg.compileRule(rt); err != nil{
            return nil, err
        } else{
            wrappedParser := pg.parser(subrule, name, rule, true)
            pg.parsers[name] = wrappedParser
            return wrappedParser, nil
        }
    default:
        return nil, &CompilerError{"Unknown rule type"}

    }

    return nil, &CompilerError{"Unknown rule encountered!"}

}


func (pg *ParserGenerator) Compile (grammar map[string]interface{}) (Parser, error) {

    pg.grammar = grammar
    pg.parsers = make(ParserMap)

    if pg.plugin == nil {
        return nil, &CompilerError{"No plugin defined!"}
    }

    pg.plugin.PrepareGrammar(grammar)

    if  _, ok := grammar["start"]; !ok{
        panic("Grammar does not contain a start rule!")
    }

    return pg.compileRule(grammar["start"])


}
