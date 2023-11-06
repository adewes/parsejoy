package main

import (
    "testing"
    "strings"
    "math/rand"
)

func BenchmarkGrammarGrammar(b *testing.B) {

	parser, _ := MakeParser[string](grammarGrammar)

	str := ""

	for i :=0;i < 10000; i++ {
		str += "baz -> bar, bar, bar, bar;"
	}

	input := MakeStringInput(str)

	b.SetBytes(int64(len(str)))

    for i := 0; i < b.N; i++ {
		if len(parser.Run(input)) != 1 {
			panic("uh oh")
		}
    }
}

func BenchmarkEGrammar(b *testing.B) {

	parser, _ := MakeParser[string](eGrammar)

	str := "b"

	for i :=0;i < 100; i++ {
		str += "+b"
	}

	input := MakeStringInput(str)

	b.SetBytes(int64(len(str)))

    for i := 0; i < b.N; i++ {
		if len(parser.Run(input)) != 1 {
			panic("uh oh")
		}
    }
}

func BenchmarkTermGrammar(b *testing.B) {

	parser, _ := MakeParser[string](termGrammar)

	str := "1"

	for i :=0;i < 100; i++ {
		str += "+1"
	}

	input := MakeStringInput(str)

	b.SetBytes(int64(len(str)))

    for i := 0; i < b.N; i++ {
		if len(parser.Run(input)) != 1 {
			panic("uh oh")
		}
    }
}

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func BenchmarkTransitions(b *testing.B) {

	transitions := make(map[string]bool)

	for i :=0; i<100; i++ {
		l := 1+rand.Intn(10)
		r := ""
		for j:=0; j<l; j++ {
			r += string(runes[rand.Intn(len(runes))])
		}
		transitions[r] = true
	}

	str := ""

	for i:=0; i<10000; i++ {
		str += string(runes[rand.Intn(len(runes))])
	}

	b.SetBytes(int64(len(str)))

    for k := 0; k < b.N; k++ {
		i := 0
	outer:
		for {
			if i >= len(str) {
				break
			} 
			for tr, _ := range transitions {
				if strings.HasPrefix(str[i:], tr) {
					i += len(tr)
					continue outer
				}
			}
			// nothing found...
			i += 1
		}
    }
}