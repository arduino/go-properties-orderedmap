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
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestPropertiesBoardsTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "boards.txt"))

	require.NoError(t, err)

	require.Equal(t, "Processor", p.Get("menu.cpu"))
	require.Equal(t, "32256", p.Get("ethernet.upload.maximum_size"))
	require.Equal(t, "{build.usb_flags}", p.Get("robotMotor.build.extra_flags"))

	ethernet := p.SubTree("ethernet")
	require.Equal(t, "Arduino Ethernet", ethernet.Get("name"))
}

func TestPropertiesTestTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "test.txt"))

	require.NoError(t, err)
	require.Equal(t, "value = 1", p.Get("key"))

	runOS := runtime.GOOS
	if runOS == "darwin" {
		runOS = "macosx"
	}

	require.Equal(t, fmt.Sprintf("is %s", runOS), p.Get("which.os"))
}

func TestExpandPropsInStringAndMissingCheck(t *testing.T) {
	aMap := NewMap()
	aMap.Set("key1", "42")
	aMap.Set("key2", "{key1}")
	aMap.Set("key3", "{key4}")

	require.Equal(t, "42 == 42 == true", aMap.ExpandPropsInString("{key1} == {key2} == true"))

	require.False(t, aMap.IsPropertyMissingInExpandPropsInString("key3", "{key1} == {key2} == true"))
	require.False(t, aMap.IsPropertyMissingInExpandPropsInString("key1", "{key1} == {key2} == true"))
	require.True(t, aMap.IsPropertyMissingInExpandPropsInString("key4", "{key4} == {key2}"))
	require.True(t, aMap.IsPropertyMissingInExpandPropsInString("key4", "{key3} == {key2}"))
}

func TestExpandPropsInString2(t *testing.T) {
	p := NewMap()
	p.Set("key2", "{key2}")
	p.Set("key1", "42")

	str := "{key1} == {key2} == true"

	str = p.ExpandPropsInString(str)
	require.Equal(t, "42 == {key2} == true", str)
}

func TestDeleteUnexpandedPropsFromString(t *testing.T) {
	p := NewMap()
	p.Set("key1", "42")
	p.Set("key2", "{key1}")

	str := "{key1} == {key2} == {key3} == true"

	str = p.ExpandPropsInString(str)
	str = DeleteUnexpandedPropsFromString(str)
	require.Equal(t, "42 == 42 ==  == true", str)
}

func TestDeleteUnexpandedPropsFromString2(t *testing.T) {
	p := NewMap()
	p.Set("key2", "42")

	str := "{key1} == {key2} == {key3} == true"

	str = p.ExpandPropsInString(str)
	str = DeleteUnexpandedPropsFromString(str)
	require.Equal(t, " == 42 ==  == true", str)
}

func TestPropertiesRedBeearLabBoardsTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "redbearlab_boards.txt"))

	require.NoError(t, err)

	require.Equal(t, 83, p.Size())
	require.Equal(t, "Blend", p.Get("blend.name"))
	require.Equal(t, "arduino:arduino", p.Get("blend.build.core"))
	require.Equal(t, "0x2404", p.Get("blendmicro16.pid.0"))

	ethernet := p.SubTree("blend")
	require.Equal(t, "arduino:arduino", ethernet.Get("build.core"))
}

func TestLoadFromBytes(t *testing.T) {
	data := []byte(`
yun.vid.0=0x2341
yun.pid.0=0x0041
yun.vid.1=0x2341
yun.pid.1=0x8041
yun.upload.tool=avrdude
yun.upload.protocol=avr109
yun.upload.maximum_size=28672
yun.upload.maximum_data_size=2560
yun.upload.speed=57600
yun.upload.disable_flushing=true
yun.upload.use_1200bps_touch=true
yun.upload.wait_for_upload_port=true
`)
	m, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.Equal(t, "57600", m.Get("yun.upload.speed"))

	data2 := []byte(`
yun.vid.0=0x2341
yun.pid.1
yun.upload.tool=avrdude
`)
	m2, err2 := LoadFromBytes(data2)
	fmt.Println(err2)
	require.Error(t, err2)
	require.Nil(t, m2)
}

func TestLoadFromSlice(t *testing.T) {
	data := []string{"yun.vid.0=0x2341",
		"yun.pid.0=0x0041",
		"yun.vid.1=0x2341",
		"yun.pid.1=0x8041",
		"yun.upload.tool=avrdude",
		"yun.upload.protocol=avr109",
		"yun.upload.maximum_size=28672",
		"yun.upload.maximum_data_size=2560",
		"yun.upload.speed=57600",
		"yun.upload.disable_flushing=true",
		"yun.upload.use_1200bps_touch=true",
		"yun.upload.wait_for_upload_port=true",
	}
	m, err := LoadFromSlice(data)
	require.NoError(t, err)
	require.Equal(t, "57600", m.Get("yun.upload.speed"))

	data2 := []string{
		"yun.vid.0=0x2341",
		"yun.pid.1",
		"yun.upload.tool=avrdude",
	}

	m2, err2 := LoadFromSlice(data2)
	fmt.Println(err2)
	require.Error(t, err2)
	require.Nil(t, m2)
}
func TestSubTreeForMultipleDots(t *testing.T) {
	p := NewMap()
	p.Set("root.lev1.prop", "hi")
	p.Set("root.lev1.prop2", "how")
	p.Set("root.lev1.prop3", "are")
	p.Set("root.lev1.prop4", "you")
	p.Set("root.lev1", "A")

	lev1 := p.SubTree("root.lev1")
	require.Equal(t, "you", lev1.Get("prop4"))
	require.Equal(t, "hi", lev1.Get("prop"))
	require.Equal(t, "how", lev1.Get("prop2"))
	require.Equal(t, "are", lev1.Get("prop3"))
}

func TestPropertiesBroken(t *testing.T) {
	_, err := Load(filepath.Join("testdata", "broken.txt"))

	require.Error(t, err)
}

func TestGetSetBoolean(t *testing.T) {
	m := NewMap()
	m.Set("a", "true")
	m.Set("b", "false")
	m.Set("c", "hello")
	m.SetBoolean("e", true)
	m.SetBoolean("f", false)

	require.True(t, m.GetBoolean("a"))
	require.False(t, m.GetBoolean("b"))
	require.False(t, m.GetBoolean("c"))
	require.False(t, m.GetBoolean("d"))
	require.True(t, m.GetBoolean("e"))
	require.False(t, m.GetBoolean("f"))
	require.Equal(t, "true", m.Get("e"))
	require.Equal(t, "false", m.Get("f"))
}

func TestKeysMethod(t *testing.T) {
	m := NewMap()
	m.Set("k1", "value")
	m.Set("k2", "othervalue")
	m.Set("k3.k4", "anothevalue")
	m.Set("k5", "value")

	k := m.Keys()
	sort.Strings(k)
	require.Equal(t, "[k1 k2 k3.k4 k5]", fmt.Sprintf("%s", k))

	v := m.Values()
	sort.Strings(v)
	require.Equal(t, "[anothevalue othervalue value value]", fmt.Sprintf("%s", v))
}

func TestEqualsAndContains(t *testing.T) {
	x := NewMap()
	x.Set("k1", "value")
	x.Set("k2", "othervalue")
	x.Set("k3.k4", "anothevalue")
	x.Set("k5", "value")

	y := NewMap()
	y.Set("k1", "value")
	y.Set("k2", "othervalue")
	y.Set("k3.k4", "anothevalue")
	y.Set("k5", "value")

	z := NewMap()
	z.Set("k2", "othervalue")
	z.Set("k1", "value")
	z.Set("k3.k4", "anothevalue")
	z.Set("k5", "value")

	require.True(t, x.ContainsKey("k1"))
	require.True(t, x.ContainsKey("k2"))
	require.True(t, x.ContainsKey("k3.k4"))
	require.True(t, x.ContainsKey("k5"))
	require.False(t, x.ContainsKey("k3"))
	require.False(t, x.ContainsKey("k4"))
	require.True(t, x.ContainsValue("value"))
	require.True(t, x.ContainsValue("othervalue"))
	require.True(t, x.ContainsValue("anothevalue"))
	require.False(t, x.ContainsValue("vvvalue"))

	require.True(t, x.Equals(y))
	require.True(t, y.Equals(x))
	require.True(t, x.Equals(z))
	require.True(t, z.Equals(x))

	require.True(t, x.EqualsWithOrder(y))
	require.True(t, y.EqualsWithOrder(x))
	require.False(t, x.EqualsWithOrder(z))
	require.False(t, z.EqualsWithOrder(x))

	data, err := paths.New("testdata/build.json").ReadFile()
	require.NoError(t, err)
	data2, err := paths.New("testdata/build-2.json").ReadFile()
	require.NoError(t, err)

	var opts *Map
	var prevOpts *Map
	json.Unmarshal([]byte(data), &opts)
	json.Unmarshal([]byte(data2), &prevOpts)
	require.False(t, opts.Equals(prevOpts))
}

func TestExtractSubIndexSets(t *testing.T) {
	data := map[string]string{
		"uno.upload_port.vid":       "0x1000",
		"uno.upload_port.pid":       "0x2000",
		"due.upload_port.0.vid":     "0x1000",
		"due.upload_port.0.pid":     "0x2000",
		"due.upload_port.1.vid":     "0x1001",
		"due.upload_port.1.pid":     "0x2001",
		"tre.upload_port.1.vid":     "0x1001",
		"tre.upload_port.1.pid":     "0x2001",
		"tre.upload_port.2.vid":     "0x1002",
		"tre.upload_port.2.pid":     "0x2002",
		"quattro.upload_port.vid":   "0x1001",
		"quattro.upload_port.pid":   "0x2001",
		"quattro.upload_port.1.vid": "0x1002",
		"quattro.upload_port.1.pid": "0x2002",
		"quattro.upload_port.2.vid": "0x1003",
		"quattro.upload_port.2.pid": "0x2003",
	}
	m := NewFromHashmap(data)

	s1 := m.ExtractSubIndexSets("uno.upload_port")
	require.Len(t, s1, 1)
	require.Equal(t, s1[0].Get("vid"), "0x1000")
	require.Equal(t, s1[0].Get("pid"), "0x2000")

	s2 := m.ExtractSubIndexSets("due.upload_port")
	require.Len(t, s2, 2)
	require.Equal(t, s2[0].Get("vid"), "0x1000")
	require.Equal(t, s2[0].Get("pid"), "0x2000")
	require.Equal(t, s2[1].Get("vid"), "0x1001")
	require.Equal(t, s2[1].Get("pid"), "0x2001")

	s3 := m.ExtractSubIndexSets("tre.upload_port")
	require.Len(t, s3, 2)
	require.Equal(t, s3[0].Get("vid"), "0x1001")
	require.Equal(t, s3[0].Get("pid"), "0x2001")
	require.Equal(t, s3[1].Get("vid"), "0x1002")
	require.Equal(t, s3[1].Get("pid"), "0x2002")

	s4 := m.ExtractSubIndexSets("quattro.upload_port")
	require.Len(t, s4, 2)
	require.Equal(t, s4[0].Get("vid"), "0x1002")
	require.Equal(t, s4[0].Get("pid"), "0x2002")
	require.Equal(t, s4[1].Get("vid"), "0x1003")
	require.Equal(t, s4[1].Get("pid"), "0x2003")
}

func TestExtractSubIndexLists(t *testing.T) {
	data := map[string]string{
		"uno.discovery.required":       "item",
		"due.discovery.required.0":     "item1",
		"due.discovery.required.1":     "item2",
		"due.discovery.required.2":     "item3",
		"tre.discovery.required.1":     "itemA",
		"tre.discovery.required.2":     "itemB",
		"tre.discovery.required.3":     "itemC",
		"quattro.discovery.required":   "itemA",
		"quattro.discovery.required.1": "itemB",
		"quattro.discovery.required.2": "itemC",
		"cinque.discovery.something":   "itemX",
		"sei.discovery.something.1":    "itemA",
		"sei.discovery.something.2":    "itemB",
		"sei.discovery.something.5":    "itemC",
		"sei.discovery.something.12":   "itemD",
		"sette.discovery.something.01": "itemA",
		"sette.discovery.something.2":  "itemB",
		"sette.discovery.something.05": "itemC",
		"sette.discovery.something.5":  "itemD",
	}
	m := NewFromHashmap(data)

	require.Equal(t, []string{"item"}, m.ExtractSubIndexLists("uno.discovery.required"))
	require.Equal(t, []string{"item1", "item2", "item3"}, m.ExtractSubIndexLists("due.discovery.required"))
	require.Equal(t, []string{"itemA", "itemB", "itemC"}, m.ExtractSubIndexLists("tre.discovery.required"))
	require.Equal(t, []string{"itemB", "itemC"}, m.ExtractSubIndexLists("quattro.discovery.required"))
	require.Equal(t, []string{}, m.ExtractSubIndexLists("cinque.discovery.required"))
	require.Equal(t, []string{"itemA", "itemB", "itemC", "itemD"}, m.ExtractSubIndexLists("sei.discovery.something"))
	require.Equal(t, []string{"itemB", "itemD"}, m.ExtractSubIndexLists("sette.discovery.something"))
}

func TestLoadingNonUTF8Properties(t *testing.T) {
	m, err := LoadFromPath(paths.New("testdata/non-utf8.properties"))
	require.NoError(t, err)
	require.Equal(t, "Aáa", m.Get("maintainer"))
}

func TestAsSlice(t *testing.T) {
	emptyProperties := NewMap()
	require.Len(t, emptyProperties.AsSlice(), 0)

	properties := NewMap()
	properties.Set("key1", "value1")
	properties.Set("key2", "value2")
	properties.Set("key3", "value3=somethingElse")

	require.Len(t, properties.AsSlice(), properties.Size())

	require.Equal(t, []string{
		"key1=value1",
		"key2=value2",
		"key3=value3=somethingElse"},
		properties.AsSlice())
}
