// This file is part of go-properties-orderedmap library.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-2.0-or-later

package properties

import (
	"strings"

	"github.com/arduino/go-paths-helper"
)

// GetBoolean returns true if the map contains the specified key and the value
// equals to the string "true", in any other case returns false.
func (m *Map) GetBoolean(key string) bool {
	value, ok := m.GetOk(key)
	return ok && strings.TrimSpace(value) == "true"
}

// SetBoolean sets the specified key to the string "true" or "false" if the value
// is respectively true or false.
func (m *Map) SetBoolean(key string, value bool) {
	if value {
		m.Set(key, "true")
	} else {
		m.Set(key, "false")
	}
}

// GetPath returns a paths.Path object using the map value as path. The function
// returns nil if the key is not present.
func (m *Map) GetPath(key string) *paths.Path {
	value, ok := m.GetOk(key)
	if !ok {
		return nil
	}
	return paths.New(value)
}

// SetPath saves the paths.Path object in the map using the path as value of the map
func (m *Map) SetPath(key string, value *paths.Path) {
	if value == nil {
		m.Set(key, "")
	} else {
		m.Set(key, value.String())
	}
}
