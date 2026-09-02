// This file is part of go-properties-orderedmap library.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-2.0-or-later

package properties

import (
	"encoding/json"
)

// XXX: no simple way to preserve ordering in JSON.

// MarshalJSON implements json.Marshaler interface
func (m *Map) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.kv)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *Map) UnmarshalJSON(d []byte) error {
	var obj map[string]string
	if err := json.Unmarshal(d, &obj); err != nil {
		return err
	}

	m.kv = map[string]string{}
	m.o = []string{}
	for k, v := range obj {
		m.Set(k, v)
	}
	return nil
}
