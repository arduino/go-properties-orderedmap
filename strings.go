// This file is part of go-properties-orderedmap library.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-2.0-or-later

package properties

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// SplitQuotedString splits a string by spaces and at the same time allows
// to use spaces in a single element of the split using quote characters.
//
// For example the call:
//
//	SplitQuotedString(`This 'is an' "Hello World!" example`, `'"`, false)
//
// returns the following array:
//
//	[]string{"This", "is an", "Hello World!", "example"}
//
// The quoteChars parameter is a string containing all the characters that
// are considered as quote characters. If a quote character is found, the
// function will consider the text between the quote character and the next
// quote character as a single element of the split.
//
// The acceptEmptyArguments parameter is a boolean that indicates if the
// function should consider empty arguments as valid elements of the split.
// If set to false, the function will ignore empty arguments.
//
// If the function finds an opening quote character and does not find the
// closing quote character, it will return an error. In any case, the function
// will return the split array up to the point where the error occurred.
//
// The function does not support escaping of quote characters.
//
// The function is UTF-8 safe.
func SplitQuotedString(src string, quoteChars string, acceptEmptyArguments bool) ([]string, error) {
	// Make a map of valid quote runes
	isQuote := map[rune]bool{}
	for _, c := range quoteChars {
		isQuote[c] = true
	}

	result := []string{}

	var escapingChar rune
	escapedArg := ""

	for _, current := range strings.Split(src, " ") {
		if escapingChar == 0 {
			first, size := firstRune(current)
			if !isQuote[first] {
				if acceptEmptyArguments || len(strings.TrimSpace(current)) > 0 {
					result = append(result, current)
				}
				continue
			}

			escapingChar = first
			current = current[size:]
			escapedArg = ""
		}

		last, size := lastRune(current)
		if last != escapingChar {
			escapedArg += current + " "
			continue
		}

		escapedArg += current[:len(current)-size]
		if acceptEmptyArguments || len(strings.TrimSpace(escapedArg)) > 0 {
			result = append(result, escapedArg)
		}
		escapingChar = 0
	}

	if escapingChar != 0 {
		return result, fmt.Errorf("invalid quoting, no closing `%c` char found", escapingChar)
	}

	return result, nil
}

func firstRune(s string) (rune, int) {
	if len(s) == 0 || !utf8.ValidString(s) {
		return 0, 0
	}
	return utf8.DecodeRuneInString(s)
}

func lastRune(s string) (rune, int) {
	if len(s) == 0 || !utf8.ValidString(s) {
		return 0, 0
	}
	return utf8.DecodeLastRuneInString(s)
}
