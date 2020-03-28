package main

import (
	"testing"
)

type testStruct struct {
	Test string
}

func TestStringify(t *testing.T) {
	test := testStruct{
		Test: "variable",
	}

	expectedResult := "{\n  \"Test\": \"variable\"\n}"

	if stringify(test) != expectedResult {
		t.Error("Did not stringify interface correctly: " + stringify(test))
	}
}
