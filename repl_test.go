package main

import (
	"fmt"
	"testing"
	"time"
)

var pass int
var fail int

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
			input:    "",
			expected: make([]string, 1),
		},
	}
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
	}
}

func TestCacheAddGet(t *testing.T) {
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com/",
			val: []byte("test data"),
		},
		{
			key: "https://example.com/?query=test&query2=test2",
			val: []byte("test data from query"),
		},
	}
	const interval = 5 * time.Millisecond
	for _, c := range cases {
		testCache := NewCache(interval)

		testCache.Add(c.key, c.val)
		val, ok := testCache.Get(c.key)
		if !ok {
			fail++
			t.Errorf("expected to find key")
			return
		}
		if string(val) != string(c.val) {
			fail++
			t.Error("expected to find value")
			return
		}
		pass++
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	_, ok := cache.Get("https://example.com")

	if !ok {
		fail++
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime) // maybe i can wrap this in a go func later? sleeping the whole test suite seems cringe

	_, ok = cache.Get("https://example.com")
	if ok {
		fail++
		t.Errorf("expected to not find key")
		return
	}
	pass++
}

func TestPrintPassFail(t *testing.T) {
	fmt.Printf("-----------------\npass: %v, fail: %v\n", pass, fail)
}
