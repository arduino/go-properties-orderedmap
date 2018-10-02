/*
 * This file is part of PropertiesOrderedMap library.
 *
 * Copyright 2017-2018 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesOrderedMap library is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 */

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
	_, err := SplitQuotedString(`'this is' a 'test of quoting`, `"'`, true)
	require.Error(t, err)
	_, err = SplitQuotedString(`'this is' a "'test" of "quoting`, `"'`, true)
	require.Error(t, err)
}
