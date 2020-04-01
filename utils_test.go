package main

import (
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, expectedResult, stringify(test), "Did not stringify interface correctly: "+stringify(test))
}
