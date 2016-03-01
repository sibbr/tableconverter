package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestMelt We test for trim, quotes and allowing \n inside quotes
func TestMelt(t *testing.T) {
	input :=
		`        label1         ,label2,"label 3"
1,2," 3
"
"f1","f2","f3"
`
	output :=
		`eventid,label2,measurementType,measurementValue
1,2,label1,1
1,2,label 3," 3
"
2,f2,label1,f1
2,f2,label 3,f3
`
	reader := strings.NewReader(input)
	var writer bytes.Buffer
	err := Melt(reader, &writer, []string{"label2"}, ",")
	if err != nil {
		t.Error(err)
	}
	if writer.String() != output {
		t.Errorf("Input:\n%s\nGot:\n%s\nExpected:\n%s\n", input, writer.String(), output)
	}
}

func TestContains(t *testing.T) {
	inputList := []string{"test1", "test2", "test3"}
	inputSearch := "test2"
	expected := true
	output := contains(inputSearch, &inputList)
	if expected != output {
		t.Errorf("Inputs: find %q at %q\nGot: %v\nExpected: %v\n", inputSearch,
			inputList, output, expected)
	}
}

func TestIndexContains(t *testing.T) {
	inputList := []string{"test1", "test2", "test3"}
	inputSearch := "test2"
	expected := 1
	output := indexContains(inputSearch, &inputList)
	if expected != output {
		t.Errorf("Inputs: find %q at %q\nGot: %v\nExpected: %v\n", inputSearch,
			inputList, output, expected)
	}
}
