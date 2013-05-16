// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package walk

import (
	"bytes"
	"math/big"
	"strconv"
	"strings"
	"syscall"
	"time"
)

import (
	. "github.com/lxn/go-winapi"
)

var (
	decimalSepB      byte
	decimalSepUint16 uint16
	decimalSepS      string
	groupSepB        byte
	groupSepUint16   uint16
	groupSepS        string
)

func init() {
	sPtr := syscall.StringToUTF16Ptr("1000.00")

	var buf [9]uint16

	if 0 == GetNumberFormat(
		LOCALE_USER_DEFAULT,
		0,
		sPtr,
		nil,
		&buf[0],
		int32(len(buf))) {

		panic("GetNumberFormat")
	}

	s := syscall.UTF16ToString(buf[:])

	decimalSepB = s[5]
	decimalSepUint16 = buf[5]
	decimalSepS = s[5:6]
	groupSepB = s[1]
	groupSepUint16 = buf[1]
	groupSepS = s[1:2]
}

func maxi(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func mini(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

func uint16IndexUint16(s []uint16, v uint16) int {
	for i, u := range s {
		if u == v {
			return i
		}
	}

	return -1
}

func uint16ContainsUint16(s []uint16, v uint16) bool {
	return uint16IndexUint16(s, v) != -1
}

func uint16CountUint16(s []uint16, v uint16) int {
	var count int

	for _, u := range s {
		if u == v {
			count++
		}
	}

	return count
}

func uint16RemoveUint16(s []uint16, v uint16) []uint16 {
	count := uint16CountUint16(s, v)
	if count == 0 {
		return s
	}

	ret := make([]uint16, 0, len(s)-count)

	for _, u := range s {
		if u != v {
			ret = append(ret, u)
		}
	}

	return ret
}

func ParseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)

	t := FormatFloatGrouped(1000, 2)

	replaceSep := func(new string, index func(string, func(rune) bool) int) {
		i := index(t, func(r rune) bool {
			return r < '0' || r > '9'
		})

		var sep string
		if i > -1 {
			sep = string(t[i])
		}
		if sep != "" {
			s = strings.Replace(s, string(sep), new, -1)
		}
	}

	replaceSep("", strings.IndexFunc)
	replaceSep(".", strings.LastIndexFunc)

	return strconv.ParseFloat(s, 64)
}

func FormatFloat(f float64, prec int) string {
	return formatFloatString(strconv.FormatFloat(f, 'f', prec, 64), prec, false)
}

func FormatFloatGrouped(f float64, prec int) string {
	return formatFloatString(strconv.FormatFloat(f, 'f', prec, 64), prec, true)
}

func formatBigRat(r *big.Rat, prec int) string {
	return formatFloatString(r.FloatString(prec), prec, false)
}

func formatBigRatGrouped(r *big.Rat, prec int) string {
	return formatFloatString(r.FloatString(prec), prec, true)
}

func formatFloatString(s string, prec int, grouped bool) string {
	switch s {
	case "NaN", "-Inf", "+Inf":
		return s
	}

	s = strings.Replace(s, ".", decimalSepS, 1)
	if !grouped {
		return s
	}

	b := new(bytes.Buffer)

	var firstDigit int
	if len(s) > 0 && s[0] == '-' {
		firstDigit = 1
		b.WriteByte('-')
		s = s[1:]
	}

	intLen := len(s) - prec - 1

	n := intLen % 3
	if n != 0 {
		b.WriteString(s[:n])
	}
	for i := n; i < intLen; i += 3 {
		if b.Len() > firstDigit {
			b.WriteByte(groupSepB)
		}
		b.WriteString(s[i : i+3])
	}

	b.WriteString(s[intLen:])

	return b.String()
}

func setDescendantsEnabled(widget Widget, enabled bool) {
	wb := widget.BaseWidget()
	wb.SetEnabled(enabled)

	walkDescendants(widget, func(w Widget) bool {
		if w.Handle() == wb.hWnd {
			return true
		}

		EnableWindow(w.Handle(), enabled && w.BaseWidget().enabled)

		return true
	})
}

func setDescendantsFont(widget Widget, f *Font) {
	wb := widget.BaseWidget()
	wb.SetFont(f)

	walkDescendants(widget, func(w Widget) bool {
		if w.Handle() == wb.hWnd {
			return true
		}

		if w.BaseWidget().font != nil {
			return false
		}

		setWidgetFont(w.Handle(), f)

		return true
	})
}

func walkDescendants(widget Widget, f func(w Widget) bool) {
	if widget == nil || !f(widget) {
		return
	}

	var children []Widget

	switch w := widget.(type) {
	case *NumberEdit:
		children = append(children, w.edit)

	case *TabWidget:
		for _, p := range w.Pages().items {
			children = append(children, p)
		}

	case Container:
		if c := w.Children(); c != nil {
			children = c.items
		} else {
			children = nil
		}
	}

	for _, w := range children {
		walkDescendants(w, f)
	}
}

func less(a, b interface{}, order SortOrder) bool {
	c := func(ls bool) bool {
		if order == SortAscending {
			return ls
		}

		return !ls
	}

	if _, ok := a.(error); ok {
		_, bIsErr := b.(error)

		return c(!bIsErr)
	}
	if _, ok := b.(error); ok {
		return c(false)
	}

	if a == nil {
		return c(b != nil)
	}
	if b == nil {
		return c(false)
	}

	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return c(av < bv)
		}

	case int:
		if bv, ok := b.(int); ok {
			return c(av < bv)
		}

	case float64:
		if bv, ok := b.(float64); ok {
			return c(av < bv)
		}

	case float32:
		if bv, ok := b.(float32); ok {
			return c(av < bv)
		}

	case int64:
		if bv, ok := b.(int64); ok {
			return c(av < bv)
		}

	case int32:
		if bv, ok := b.(int32); ok {
			return c(av < bv)
		}

	case int16:
		if bv, ok := b.(int16); ok {
			return c(av < bv)
		}

	case int8:
		if bv, ok := b.(int8); ok {
			return c(av < bv)
		}

	case uint:
		if bv, ok := b.(uint); ok {
			return c(av < bv)
		}

	case uint64:
		if bv, ok := b.(uint64); ok {
			return c(av < bv)
		}

	case uint32:
		if bv, ok := b.(uint32); ok {
			return c(av < bv)
		}

	case uint16:
		if bv, ok := b.(uint16); ok {
			return c(av < bv)
		}

	case uint8:
		if bv, ok := b.(uint8); ok {
			return c(av < bv)
		}

	case time.Time:
		if bv, ok := b.(time.Time); ok {
			return c(av.Before(bv))
		}

	case bool:
		if bv, ok := b.(bool); ok {
			return c(!av && bv)
		}
	}

	return false
}
