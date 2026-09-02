// This file is part of go-properties-orderedmap library.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-2.0-or-later

package properties

import (
	"fmt"
	"strings"
)

// DebugExpandPropsInString outputs the substitutions made by
// ExpandPropsInString for debugging purposes.
func (m *Map) DebugExpandPropsInString(str string) string {
	return m.expandProps(str, true)
}

func (m *Map) expandProps(str string, debug bool) string {
	debug = debug || m.Debug
	for i := 0; i < 10; i++ {
		if debug {
			fmt.Printf("pass %d: %s\n", i, str)
		}
		newStr := str
		for key, value := range m.kv {
			if debug && strings.Contains(newStr, "{"+key+"}") {
				fmt.Printf("  Replacing %s -> %s\n", key, value)
			}
			newStr = strings.Replace(newStr, "{"+key+"}", value, -1)
		}
		if str == newStr {
			break
		}
		str = newStr
	}
	return str
}

// Dump returns a representation of the map in golang source format
func (m *Map) Dump() string {
	res := "properties.Map{\n"
	for _, k := range m.o {
		res += fmt.Sprintf("  \"%s\": \"%s\",\n", strings.Replace(k, `"`, `\"`, -1), strings.Replace(m.Get(k), `"`, `\"`, -1))
	}
	res += "}"
	return res
}
