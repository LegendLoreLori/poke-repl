package main

import (
	"fmt"
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "   regirock registeel  ",
			expected: []string{"regirock", "registeel"},
		},
		{
			input:    "pikachu bulbasaur espurr whismurr noivern crustle",
			expected: []string{"pikachu", "bulbasaur", "espurr", "whismurr", "noivern", "crustle"},
		},
		{
			input:    "PIKACHU",
			expected: []string{"pikachu"},
		},
		{
			input:    "ch22armander",
			expected: []string{"TODO: implement err handling"},
		},
		{
			input:    "",
			expected: make([]string, 1),
		},
		{
			input:    "Agumon",
			expected: []string{"TODO: implement err handling"},
		},
	}
	pass := 0
	fail := 0
	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			fail++
			t.Errorf("-----------------\ninput: '%v'\nexpected: %v\nactual: %v\n", c.input, len(c.expected), len(actual))
		}
		for i := 0; i < len(actual); i++ {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				fail++
				t.Errorf("-----------------\ninput: '%v'\nexpected: %v\nactual: %v\n", c.input, c.expected, actual)
				break
			}
		}
		pass++
		fmt.Printf("-----------------\ninput: '%v'\nexpected: %v\nactual: %v\n", c.input, c.expected, actual)
	}
	fmt.Printf("-----------------\npass: %v, fail: %v\n", pass, fail)
}
