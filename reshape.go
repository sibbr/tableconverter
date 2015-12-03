package main

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
)

// Melt will change format to wide -> long
func Melt(input io.Reader, output io.Writer, fixed []string, sep string) error {

	dados := csv.NewReader(input)
	if sep == "tab" {
		dados.Comma = '\t'
	} else {
		dados.Comma = rune(sep[0])
	}
	dados.FieldsPerRecord = -1
	dados.LazyQuotes = true

	labels, err := dados.Read()

	if err != nil {
		return err
	}
	// data cleaning, removing all leading and trailing white space
	for k, v := range labels {
		labels[k] = strings.TrimSpace(v)
	}
	for k, v := range fixed {
		fixed[k] = strings.TrimSpace(v)
	}

	// stop if duplicate labels are found
	found := map[string]int{}
	anyDuplicate := []string{}
	for _, v := range labels {
		if _, ok := found[v]; ok {
			anyDuplicate = append(anyDuplicate, v)
		}
		found[v]++
	}
	if len(anyDuplicate) > 0 {
		return &reshapeError{"Duplicated column names: " + strings.Join(anyDuplicate, ", ")}
	}

	writeMeasurementData := csv.NewWriter(output)
	outputLabels := []string{"eventid"}
	outputLabels = append(outputLabels, fixed...)
	outputLabels = append(outputLabels, "measurementType", "measurementValue")
	if writeMeasurementData.Write(outputLabels) != nil {
		return err
	}

	fixedPos := []int{}
	for k, v := range labels {
		if indexContains(v, &fixed) > -1 {
			fixedPos = append(fixedPos, k)
		}
	}
	if len(fixedPos) < len(fixed) {
		return &reshapeError{"Fixed column not found in dataset"}
	}

	// for each line do a rotation and write, no waste in memory
	// number of resulting lines = (ncol - fixed) * nrow
	// fixed are columns like eventid to control the rotation of data
	for eventid := 1; ; eventid++ {

		line, err := dados.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		for elem := 0; elem < len(line); elem++ {
			if contains(labels[elem], &fixed) {
				continue
			}
			outputLine := []string{strconv.Itoa(eventid)}
			for _, v := range fixedPos {
				outputLine = append(outputLine, line[v])
			}
			outputLine = append(outputLine, labels[elem], line[elem])
			if writeMeasurementData.Write(outputLine) != nil {
				return err
			}
		}
	}
	writeMeasurementData.Flush()

	return nil
}

func contains(element string, elements *[]string) bool {
	for _, v := range *elements {
		if element == v {
			return true
		}
	}
	return false
}

func indexContains(element string, elements *[]string) int {
	for k, v := range *elements {
		if element == v {
			return k
		}
	}
	return -1
}
