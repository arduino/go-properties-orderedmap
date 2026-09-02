// This file is part of go-properties-orderedmap library.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-2.0-or-later

package properties

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitQuotedString(t *testing.T) {
	res, err := SplitQuotedString(`this is a "test of quoting" another test`, `"`, true)
	require.NoError(t, err)
	require.EqualValues(t, res, []string{"this", "is", "a", "test of quoting", "another", "test"})
}

func TestSplitQuotedStringMixedQuotes(t *testing.T) {
	res, err := SplitQuotedString(`this is a "test 'of' quoting" 'another test' "that's it"`, `"'`, true)
	require.NoError(t, err)
	require.EqualValues(t, res, []string{"this", "is", "a", "test 'of' quoting", "another test", "that's it"})
}

func TestSplitQuotedStringEmptyArgsAllowed(t *testing.T) {
	res, err := SplitQuotedString(`this   is  a " test 'of' quoting " `, `"'`, true)
	require.NoError(t, err)
	require.EqualValues(t, res, []string{"this", "", "", "is", "", "a", " test 'of' quoting ", ""})

	res, err = SplitQuotedString(`this   is  a " test 'of' quoting " `, `"'`, false)
	require.NoError(t, err)
	require.EqualValues(t, res, []string{"this", "is", "a", " test 'of' quoting "})
}

func TestSplitQuotedStringWithUTF8(t *testing.T) {
	res, err := SplitQuotedString(`èthis is a testè of quoting`, `è`, true)
	require.NoError(t, err)
	require.EqualValues(t, res, []string{"this is a test", "of", "quoting"})
}

func TestSplitQuotedStringInvalid(t *testing.T) {
	res, err := SplitQuotedString(`'this is' a 'test of quoting`, `"'`, true)
	require.EqualError(t, err, "invalid quoting, no closing `'` char found")
	require.Equal(t, res, []string{"this is", "a"})

	res, err = SplitQuotedString(`'this is' a "'test" of "quoting`, `"'`, true)
	require.EqualError(t, err, "invalid quoting, no closing `\"` char found")
	require.Equal(t, res, []string{"this is", "a", "'test", "of"})
}
