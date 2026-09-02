// This file is part of go-properties-orderedmap library.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-2.0-or-later

package properties

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJson(t *testing.T) {
	testmap := NewMap()
	testmap.Set("a", "1")
	testmap.Set("b", "2")
	testmap.Set(`"e"`, "3")
	testmap.Set("d", "4")
	testmap.Set(`\"c"\`, "5")

	d, err := json.Marshal(testmap)
	require.NoError(t, err)
	require.Equal(t, `{"\"e\"":"3","\\\"c\"\\":"5","a":"1","b":"2","d":"4"}`, string(d))

	var decodedmap *Map
	err = json.Unmarshal(d, &decodedmap)
	require.NoError(t, err)
	require.Equal(t, 5, decodedmap.Size())
	require.Equal(t, "1", decodedmap.Get("a"))
	require.Equal(t, "2", decodedmap.Get("b"))
	require.Equal(t, "3", decodedmap.Get(`"e"`))
	require.Equal(t, "4", decodedmap.Get("d"))
	require.Equal(t, "5", decodedmap.Get(`\"c"\`))
}
