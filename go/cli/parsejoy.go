package main

import "fmt"
import "gopkg.in/yaml.v2"
import "os"
import "log"
import "7scientists.com/parsejoy"
import "io/ioutil"
import "flag"
import "time"
import "github.com/pkg/profile"
import "runtime"

func main() {

	runtime.SetCPUProfileRate(100)
	profiler := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
	defer profiler.Stop()

	inputGrammar := flag.String("grammar", "", "the grammar to use")
	inputCode := flag.String("code", "", "the code to parse")
	debugTokenizer := flag.Bool("debug-tokenizer", false, "debug mode for tokenizer")
	debugParser := flag.Bool("debug-parser", false, "debug mode for parser")

	flag.Parse()

	if *inputGrammar == "" {
		log.Fatal("You need to specify a grammar file via the --grammar flag")
		os.Exit(-1)
	}

	if *inputCode == "" {
		log.Fatal("You need to specify a code file via the --code flag")
		os.Exit(-1)
	}

	grammarString, err := ioutil.ReadFile(*inputGrammar)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	codeString, err := ioutil.ReadFile(*inputCode)

	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	pgTokenizer := parsejoy.ParserGenerator{}
	//we set the
	stringParserPlugin := parsejoy.StringParserPlugin{}
	stringParserPlugin.Initialize()
	pgTokenizer.SetPlugin(&stringParserPlugin)

	var result map[string]interface{}
	yamlerror := yaml.Unmarshal(grammarString, &result)
	if yamlerror != nil {
		panic("Error when parsing YAML!")
	}
	fmt.Printf("Successfully loaded YAML.\n")
	var tokenizerGrammar map[string]interface{}
	grammar := result
	useTokenizer := true
	tokenizer, ok := result["tokenizer"].(map[interface{}]interface{})
	if !ok {
		tokenizerGrammar = grammar
		useTokenizer = false
	} else {
		fmt.Println("Found tokenizer grammar!")
		tokenizerGrammar = make(map[string]interface{})
		for key, value := range tokenizer {
			if ks, ok := key.(string); ok {
				tokenizerGrammar[ks] = value
			}
		}
	}

	fmt.Println(useTokenizer)

	tokenParser, err := pgTokenizer.Compile(tokenizerGrammar)

	fo, err := os.Create("prefixes.yml")

	if err != nil {
		panic("Cannot open output file!")
	}

	defer fo.Close()

	out, err := yaml.Marshal(stringParserPlugin.RulePrefixes)

	fo.Write(out)

	if err != nil {
		panic("Error when compiling token parser")
	}

	var token parsejoy.BaseToken
	var newState parsejoy.State

	//parsejoy.PrintParseTree(stringToken,stringParserPlugin.TokenIds,0)

	pgParser := parsejoy.ParserGenerator{}
	tokenParserPlugin := parsejoy.TokenParserPlugin{}
	tokenParserPlugin.Initialize(stringParserPlugin.TokenIds)
	pgParser.SetPlugin(&tokenParserPlugin)
	parser, err := pgParser.Compile(grammar)

	if err != nil {
        fmt.Println(err)
		panic("Error when compiling parser")
	}

	var tokenizingTime int64 = 0
	var parsingTime int64 = 0
	var processingTime int64 = 0

	var newState2 parsejoy.State
	var newStringState *parsejoy.StringState
	var newToken2 parsejoy.BaseToken

	repetitions := 10
	for i := 0; i < repetitions; i++ {

		currentTime := time.Now().UnixNano()

		state := parsejoy.StringState{}
		state.Initialize(codeString, stringParserPlugin.TokenIds)
		state.Debug = *debugTokenizer
		processingTime += time.Now().UnixNano()-currentTime
		currentTime = time.Now().UnixNano()

		newState, token, err = tokenParser(&state)
		newStringState, _ = newState.(*parsejoy.StringState)
		stringToken, _ := token.(*parsejoy.Token)

		tokenizingTime += time.Now().UnixNano()-currentTime
		currentTime = time.Now().UnixNano()

		if err != nil {
			fmt.Println(err)
			fmt.Println("------------------")
			fmt.Println(string(newStringState.Value()))
			panic("Error when tokenizing input!")
		}

		currentTime = time.Now().UnixNano()

		tokenState := parsejoy.TokenState{}
		parsejoy.LinkTokens(stringToken)
		tokenState.Initialize(stringToken, &state)
		tokenState.Debug = *debugParser

		processingTime += time.Now().UnixNano()-currentTime
		currentTime = time.Now().UnixNano()

		newState2, newToken2, err = parser(&tokenState)
		if err != nil {
			fmt.Println(err)
		}
		newTokenState := newState2.(*parsejoy.TokenState)
		if newTokenState.CurrentToken != nil {
			fmt.Println(newTokenState.CurrentToken)
			panic("uh oh")
		}

		parsingTime += time.Now().UnixNano()-currentTime

	}

	_, _ = newToken2.(*parsejoy.L2Token)
	//stringToken, _ := token.(*parsejoy.Token)

	if err != nil {
		fmt.Println(err)
		panic("Error when parsing input")
	}

	newTokenState := newState2.(*parsejoy.TokenState)

	if err != nil {
		panic("")
	}

	//parsejoy.PrintL2ParseTree(l2Token,0)
	//parsejoy.PrintParseTree(stringToken,stringParserPlugin.TokenIds,0)

	fmt.Println(newTokenState.Context.Calls, "calls made")
	fmt.Printf("Parsing took %.2d ms, tokenizing took %.2d ms, processing took %.2d ms, total time: %.2d\n%.2d lines, parsing at %.2f loc/s.\n", parsingTime/1e6,tokenizingTime/1e6,processingTime/1e6,(parsingTime+tokenizingTime+processingTime)/1e6,newStringState.Context.NumberOfLines,float64(newStringState.Context.NumberOfLines)/float64(parsingTime+tokenizingTime+processingTime)*float64(repetitions)*1e9)

	fmt.Println(newStringState.Context.Calls, "calls made,", newStringState.Context.Errors, "errors")


}
