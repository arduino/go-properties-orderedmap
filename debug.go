/*
 * This file is part of PropertiesOrderedMap library.
 *
 * Copyright 2017-2018 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesMap library is free software; you can redistribute it and/or modify
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
