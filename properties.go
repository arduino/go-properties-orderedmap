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

/*
Package properties is a library for handling maps of hierarchical properties.
This library is mainly used in the Arduino platform software to handle
configurations made of key/value pairs stored in files with an INI like
syntax, for example:

	...
	uno.name=Arduino/Genuino Uno
	uno.upload.tool=avrdude
	uno.upload.protocol=arduino
	uno.upload.maximum_size=32256
	uno.upload.maximum_data_size=2048
	uno.upload.speed=115200
	uno.build.mcu=atmega328p
	uno.build.f_cpu=16000000L
	uno.build.board=AVR_UNO
	uno.build.core=arduino
	uno.build.variant=standard
	diecimila.name=Arduino Duemilanove or Diecimila
	diecimila.upload.tool=avrdude
	diecimila.upload.protocol=arduino
	diecimila.build.f_cpu=16000000L
	diecimila.build.board=AVR_DUEMILANOVE
	diecimila.build.core=arduino
	diecimila.build.variant=standard
	...

This library has methods to parse this kind of file into a Map object.

The Map internally keeps the insertion order so it can be retrieved later when
cycling through the key sets.

The Map object has many helper methods to accomplish some common operations
on this kind of data like cloning, merging, comparing and also extracting
a submap or generating a map-of-submaps from the first "level" of the hierarchy.

On the Arduino platform the properties are used to populate command line recipes
so there are some methods to help this task like SplitQuotedString or ExpandPropsInString.
*/
package properties

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/arduino/go-paths-helper"
)

// Map is a container of properties
type Map struct {
	kv map[string]string
	o  []string

	// Debug if set to true ExpandPropsInString will always output debugging information
	Debug bool
}

var osSuffix string

func init() {
	switch value := runtime.GOOS; value {
	case "darwin":
		osSuffix = "macosx"
	default:
		osSuffix = runtime.GOOS
	}
}

// GetOSSuffix returns the os name used to filter os-specific properties in Load* functions
func GetOSSuffix() string {
	return osSuffix
}

// SetOSSuffix forces the OS suffix to the given value
func SetOSSuffix(suffix string) {
	osSuffix = suffix
}

// NewMap returns a new Map
func NewMap() *Map {
	return &Map{
		kv: map[string]string{},
		o:  []string{},
	}
}

// NewFromHashmap creates a new Map from the given map[string]string.
// Insertion order is not preserved.
func NewFromHashmap(orig map[string]string) *Map {
	res := NewMap()
	for k, v := range orig {
		res.Set(k, v)
	}
	return res
}

func toUtf8(iso8859_1_buf []byte) string {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

// LoadFromBytes reads properties data and makes a Map out of it.
func LoadFromBytes(bytes []byte) (*Map, error) {
	var text string
	if utf8.Valid(bytes) {
		text = string(bytes)
	} else {
		// Assume ISO8859-1 encoding and convert to UTF-8
		text = toUtf8(bytes)
	}
	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.Replace(text, "\r", "\n", -1)

	properties := NewMap()

	for lineNum, line := range strings.Split(text, "\n") {
		if err := properties.parseLine(line); err != nil {
			return nil, fmt.Errorf("error parsing data at line %d: %s", lineNum, err)
		}
	}

	return properties, nil
}

// Load reads a properties file and makes a Map out of it.
func Load(filepath string) (*Map, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err)
	}

	res, err := LoadFromBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err)
	}
	return res, nil
}

// LoadFromPath reads a properties file and makes a Map out of it.
func LoadFromPath(path *paths.Path) (*Map, error) {
	return Load(path.String())
}

// LoadFromSlice reads a properties file from an array of string
// and makes a Map out of it
func LoadFromSlice(lines []string) (*Map, error) {
	properties := NewMap()

	for lineNum, line := range lines {
		if err := properties.parseLine(line); err != nil {
			return nil, fmt.Errorf("error reading from slice (index:%d): %s", lineNum, err)
		}
	}

	return properties, nil
}

func (m *Map) parseLine(line string) error {
	line = strings.TrimSpace(line)

	// Skip empty lines or comments
	if len(line) == 0 || line[0] == '#' {
		return nil
	}

	lineParts := strings.SplitN(line, "=", 2)
	if len(lineParts) != 2 {
		return fmt.Errorf("invalid line format, should be 'key=value'")
	}
	key := strings.TrimSpace(lineParts[0])
	value := strings.TrimSpace(lineParts[1])

	key = strings.Replace(key, "."+osSuffix, "", 1)
	m.Set(key, value)

	return nil
}

// SafeLoadFromPath is like LoadFromPath, except that it returns an empty Map if
// the specified file doesn't exist
func SafeLoadFromPath(path *paths.Path) (*Map, error) {
	return SafeLoad(path.String())
}

// SafeLoad is like Load, except that it returns an empty Map if the specified
// file doesn't exist
func SafeLoad(filepath string) (*Map, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return NewMap(), nil
	}

	properties, err := Load(filepath)
	if err != nil {
		return nil, err
	}
	return properties, nil
}

// Get retrieves the value corresponding to key
func (m *Map) Get(key string) string {
	return m.kv[key]
}

// GetOk retrieves the value corresponding to key and returns a true/false indicator
// to check if the key is present in the map (true if the key is present)
func (m *Map) GetOk(key string) (string, bool) {
	v, ok := m.kv[key]
	return v, ok
}

// ContainsKey returns true if the map contains the specified key
func (m *Map) ContainsKey(key string) bool {
	_, has := m.kv[key]
	return has
}

// ContainsValue returns true if the map contains the specified value
func (m *Map) ContainsValue(value string) bool {
	for _, v := range m.kv {
		if v == value {
			return true
		}
	}
	return false
}

// Set inserts or replaces an existing key-value pair in the map
func (m *Map) Set(key, value string) {
	if _, has := m.kv[key]; has {
		m.Remove(key)
	}
	m.kv[key] = value
	m.o = append(m.o, key)
}

// Size returns the number of elements in the map
func (m *Map) Size() int {
	return len(m.kv)
}

// Remove removes the key from the map
func (m *Map) Remove(key string) {
	delete(m.kv, key)
	for i, k := range m.o {
		if k == key {
			m.o = append(m.o[:i], m.o[i+1:]...)
			return
		}
	}
}

// FirstLevelOf generates a map-of-Maps using the first level of the hierarchy
// of the current Map. For example the following Map:
//
//	properties.Map{
//	  "uno.name": "Arduino/Genuino Uno",
//	  "uno.upload.tool": "avrdude",
//	  "uno.upload.protocol": "arduino",
//	  "uno.upload.maximum_size": "32256",
//	  "diecimila.name": "Arduino Duemilanove or Diecimila",
//	  "diecimila.upload.tool": "avrdude",
//	  "diecimila.upload.protocol": "arduino",
//	  "diecimila.bootloader.tool": "avrdude",
//	  "diecimila.bootloader.low_fuses": "0xFF",
//	}
//
// is transformed into the following map-of-Maps:
//
//	map[string]Map{
//	  "uno" : properties.Map{
//	    "name": "Arduino/Genuino Uno",
//	    "upload.tool": "avrdude",
//	    "upload.protocol": "arduino",
//	    "upload.maximum_size": "32256",
//	  },
//	  "diecimila" : properties.Map{
//	    "name": "Arduino Duemilanove or Diecimila",
//	    "upload.tool": "avrdude",
//	    "upload.protocol": "arduino",
//	    "bootloader.tool": "avrdude",
//	    "bootloader.low_fuses": "0xFF",
//	  }
//	}
func (m *Map) FirstLevelOf() map[string]*Map {
	newMap := make(map[string]*Map)
	for _, key := range m.o {
		if !strings.Contains(key, ".") {
			continue
		}
		keyParts := strings.SplitN(key, ".", 2)
		if newMap[keyParts[0]] == nil {
			newMap[keyParts[0]] = NewMap()
		}
		value := m.kv[key]
		newMap[keyParts[0]].Set(keyParts[1], value)
	}
	return newMap
}

// FirstLevelKeys returns the keys in the first level of the hierarchy
// of the current Map. For example the following Map:
//
//	properties.Map{
//	  "uno.name": "Arduino/Genuino Uno",
//	  "uno.upload.tool": "avrdude",
//	  "uno.upload.protocol": "arduino",
//	  "uno.upload.maximum_size": "32256",
//	  "diecimila.name": "Arduino Duemilanove or Diecimila",
//	  "diecimila.upload.tool": "avrdude",
//	  "diecimila.upload.protocol": "arduino",
//	  "diecimila.bootloader.tool": "avrdude",
//	  "diecimila.bootloader.low_fuses": "0xFF",
//	}
//
// will produce the following result:
//
//	[]string{
//	  "uno",
//	  "diecimila",
//	}
//
// the order of the original map is preserved
func (m *Map) FirstLevelKeys() []string {
	res := []string{}
	taken := map[string]bool{}
	for _, k := range m.o {
		first := strings.SplitN(k, ".", 2)[0]
		if taken[first] {
			continue
		}
		taken[first] = true
		res = append(res, first)
	}
	return res
}

// SubTree extracts a sub Map from an existing map using the first level
// of the keys hierarchy as selector.
// For example the following Map:
//
//	properties.Map{
//	  "uno.name": "Arduino/Genuino Uno",
//	  "uno.upload.tool": "avrdude",
//	  "uno.upload.protocol": "arduino",
//	  "uno.upload.maximum_size": "32256",
//	  "diecimila.name": "Arduino Duemilanove or Diecimila",
//	  "diecimila.upload.tool": "avrdude",
//	  "diecimila.upload.protocol": "arduino",
//	  "diecimila.bootloader.tool": "avrdude",
//	  "diecimila.bootloader.low_fuses": "0xFF",
//	}
//
// after calling SubTree("uno") will be transformed into:
//
//	properties.Map{
//	  "name": "Arduino/Genuino Uno",
//	  "upload.tool": "avrdude",
//	  "upload.protocol": "arduino",
//	  "upload.maximum_size": "32256",
//	},
func (m *Map) SubTree(rootKey string) *Map {
	rootKey += "."
	newMap := NewMap()
	for _, key := range m.o {
		if !strings.HasPrefix(key, rootKey) {
			continue
		}
		value := m.kv[key]
		newMap.Set(key[len(rootKey):], value)
	}
	return newMap
}

// ExpandPropsInString uses the Map to replace values into a format string.
// The format string should contains markers between braces, for example:
//
//	"The selected upload protocol is {upload.protocol}."
//
// Each marker is replaced by the corresponding value of the Map.
// The values in the Map may contain other markers, they are evaluated
// recursively up to 10 times.
func (m *Map) ExpandPropsInString(str string) string {
	return m.expandProps(str, false)
}

// IsPropertyMissingInExpandPropsInString checks if a property 'prop' is missing
// when the ExpandPropsInString method is applied to the input string 'str'.
// This method returns false if the 'prop' is defined in the map
// or if 'prop' is not used in the string expansion of 'str', otherwise
// the method returns true.
func (m *Map) IsPropertyMissingInExpandPropsInString(prop, str string) bool {
	if m.ContainsKey(prop) {
		return false
	}

	xm := m.Clone()

	// Find a random tag that is not contained in the dictionary and the src pattern
	var token string
	for {
		token = fmt.Sprintf("%d", rand.Int63())
		if strings.Contains(str, token) {
			continue
		}
		if xm.ContainsKey(token) {
			continue
		}
		if xm.ContainsValue(token) {
			continue
		}
		break
	}
	xm.Set(prop, token)

	res := xm.expandProps(str, false)
	return strings.Contains(res, token)
}

// Merge merges other Maps into this one. Each key/value of the merged Maps replaces
// the key/value present in the original Map.
func (m *Map) Merge(sources ...*Map) *Map {
	for _, source := range sources {
		for _, key := range source.o {
			value := source.kv[key]
			m.Set(key, value)
		}
	}
	return m
}

// Keys returns an array of the keys contained in the Map
func (m *Map) Keys() []string {
	keys := make([]string, len(m.o))
	copy(keys, m.o)
	return keys
}

// Values returns an array of the values contained in the Map. Duplicated
// values are repeated in the list accordingly.
func (m *Map) Values() []string {
	values := make([]string, len(m.o))
	for i, key := range m.o {
		values[i] = m.kv[key]
	}
	return values
}

// AsMap returns the underlying map[string]string. This is useful if you need to
// for ... range but without the requirement of the ordered elements.
func (m *Map) AsMap() map[string]string {
	return m.kv
}

// AsSlice returns the underlying map[string]string as a slice of
// strings with the pattern `{key}={value}`, maintaining the insertion order of the keys.
func (m *Map) AsSlice() []string {
	properties := make([]string, len(m.o))
	for i, key := range m.o {
		properties[i] = strings.Join([]string{key, m.kv[key]}, "=")
	}
	return properties
}

// Clone makes a copy of the Map
func (m *Map) Clone() *Map {
	clone := NewMap()
	clone.Merge(m)
	return clone
}

// Equals returns true if the current Map contains the same key/value pairs of
// the Map passed as argument, the order of insertion does not matter.
func (m *Map) Equals(other *Map) bool {
	return reflect.DeepEqual(m.kv, other.kv)
}

// EqualsWithOrder returns true if the current Map contains the same key/value pairs of
// the Map passed as argument with the same order of insertion.
func (m *Map) EqualsWithOrder(other *Map) bool {
	return reflect.DeepEqual(m.o, other.o) && reflect.DeepEqual(m.kv, other.kv)
}

// MergeMapsOfProperties merges the map-of-Maps (obtained from the method FirstLevelOf()) into the
// target map-of-Maps.
func MergeMapsOfProperties(target map[string]*Map, sources ...map[string]*Map) map[string]*Map {
	for _, source := range sources {
		for key, value := range source {
			target[key] = value
		}
	}
	return target
}

// DeleteUnexpandedPropsFromString removes all the brace markers "{xxx}" that are not expanded
// into a value using the Map.ExpandPropsInString() method.
func DeleteUnexpandedPropsFromString(str string) string {
	rxp := regexp.MustCompile(`\{.+?\}`)
	return rxp.ReplaceAllString(str, "")
}

// ExtractSubIndexSets works like SubTree but it considers also the numeric sub index in the form
// `root.N.xxx...` as separate subsets.
// For example the following Map:
//
//	properties.Map{
//	  "uno.upload_port.vid": "0x1000",
//	  "uno.upload_port.pid": "0x2000",
//	  "due.upload_port.0.vid": "0x1000",
//	  "due.upload_port.0.pid": "0x2000",
//	  "due.upload_port.1.vid": "0x1001",
//	  "due.upload_port.1.pid": "0x2001",
//	  "tre.upload_port.1.vid": "0x1001",
//	  "tre.upload_port.1.pid": "0x2001",
//	  "tre.upload_port.2.vid": "0x1002",
//	  "tre.upload_port.2.pid": "0x2002",
//	}
//
// calling ExtractSubIndexSets("uno.upload_port") returns the array:
//
//	[ properties.Map{
//	    "vid": "0x1000",
//	    "pid": "0x2000",
//	  },
//	]
//
// calling ExtractSubIndexSets("due.upload_port") returns the array:
//
//	[ properties.Map{
//	    "vid": "0x1000",
//	    "pid": "0x2000",
//	  },
//	  properties.Map{
//	    "vid": "0x1001",
//	    "pid": "0x2001",
//	  },
//	]
//
// the sub-index may start with .1 too, so calling ExtractSubIndexSets("tre.upload_port") returns:
//
//	[ properties.Map{
//	    "vid": "0x1001",
//	    "pid": "0x2001",
//	  },
//	  properties.Map{
//	    "vid": "0x1002",
//	    "pid": "0x2002",
//	  },
//	]
//
// Numeric subindex cannot be mixed with non-numeric, in that case only the numeric sub
// index sets will be returned.
func (m *Map) ExtractSubIndexSets(root string) []*Map {
	res := []*Map{}
	portIDPropsSet := m.SubTree(root)
	if portIDPropsSet.Size() == 0 {
		return res
	}

	// First check the properties with numeric sub index "root.N.xxx"
	idx := 0
	haveIndexedProperties := false
	for {
		idProps := portIDPropsSet.SubTree(fmt.Sprintf("%d", idx))
		idx++
		if idProps.Size() > 0 {
			haveIndexedProperties = true
			res = append(res, idProps)
		} else if idx > 1 {
			// Always check sub-id 0 and 1 (https://github.com/arduino/arduino-cli/issues/456)
			break
		}
	}

	// if there are no subindexed then return the whole "roox.xxx" subtree
	if !haveIndexedProperties {
		res = append(res, portIDPropsSet)
	}

	return res
}

// ExtractSubIndexLists extracts a list of arguments from a root `root.N=...`.
// For example the following Map:
//
//	properties.Map{
//	  "uno.discovery.required": "item",
//	  "due.discovery.required.0": "item1",
//	  "due.discovery.required.1": "item2",
//	  "due.discovery.required.2": "item3",
//	  "tre.discovery.required.1": "itemA",
//	  "tre.discovery.required.2": "itemB",
//	  "tre.discovery.required.3": "itemC",
//	  "quattro.discovery.required.1": "itemA",
//	  "quattro.discovery.required.4": "itemB",
//	  "quattro.discovery.required.5": "itemC",
//	}
//
// calling ExtractSubIndexLists("uno.discovery.required") returns the array:
//
//	[ "item" ]
//
// calling ExtractSubIndexLists("due.discovery.required") returns the array:
//
//	[ "item1", "item2", "item3" ]
//
// the sub-index may start with .1 too, so calling ExtractSubIndexLists("tre.discovery.required") returns:
//
//	[ "itemA", "itemB", "itemC" ]
//
// also the list may contains holes, so calling ExtractSubIndexLists("quattro.discovery.required") returns:
//
//	[ "itemA", "itemB", "itemC" ]
//
// Numeric subindex cannot be mixed with non-numeric, in that case only the numeric sub
// index sets will be returned.
func (m *Map) ExtractSubIndexLists(root string) []string {
	isNotDigit := func(in string) bool {
		for _, r := range in {
			if r < '0' || r > '9' {
				return true
			}
		}
		return false
	}

	// Extract numeric keys
	subProps := m.SubTree(root)
	indexes := []int{}
	for _, key := range subProps.o {
		if isNotDigit(key) {
			continue
		}
		if idx, err := strconv.Atoi(key); err == nil {
			indexes = append(indexes, idx)
		}
	}
	sort.Ints(indexes)

	res := []string{}
	haveIndexedProperties := false
	for i, idx := range indexes {
		if i > 0 && idx == indexes[i-1] {
			// de-duplicate cases like "05" and "5"
			continue
		}
		if v, ok := subProps.GetOk(strconv.Itoa(idx)); ok {
			haveIndexedProperties = true
			res = append(res, v)
		}
	}

	// if there are no subindexed then return the whole "roox.xxx" subtree
	if !haveIndexedProperties {
		if value, ok := m.GetOk(root); ok {
			res = append(res, value)
		}
	}

	return res
}
