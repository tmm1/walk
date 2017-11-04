// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package walk

import (
	"github.com/lxn/win"
)

type Label struct {
	WidgetBase
	textChangedPublisher EventPublisher
	textColor            Color
}

func NewLabel(parent Container) (*Label, error) {
	l := new(Label)

	if err := InitWidget(
		l,
		parent,
		"STATIC",
		win.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	l.SetBackground(nullBrushSingleton)

	l.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return l.Text()
		},
		func(v interface{}) error {
			return l.SetText(v.(string))
		},
		l.textChangedPublisher.Event()))

	return l, nil
}

func (*Label) LayoutFlags() LayoutFlags {
	return GrowableVert | GrowableHorz | GreedyHorz
}

func (l *Label) MinSizeHint() Size {
	return l.calculateTextSize()
}

func (l *Label) SizeHint() Size {
	return l.MinSizeHint()
}

func (l *Label) Text() string {
	return windowText(l.hWnd)
}

func (l *Label) SetText(value string) error {
	if value == l.Text() {
		return nil
	}

	if err := setWindowText(l.hWnd, value); err != nil {
		return err
	}

	return l.updateParentLayout()
}

func (l *Label) TextColor() Color {
	return l.textColor
}

func (l *Label) SetTextColor(c Color) {
	l.textColor = c

	l.Invalidate()
}

func (l *Label) Alignment() Alignment1D {
	switch win.GetWindowLong(l.hWnd, win.GWL_STYLE) & (win.SS_LEFT | win.SS_CENTER | win.SS_RIGHT) {
	case win.SS_CENTER:
		return AlignCenter

	case win.SS_RIGHT:
		return AlignFar
	}

	return AlignNear
}

func (l *Label) SetAlignment(alignment Alignment1D) error {
	var bit uint32

	switch alignment {
	case AlignCenter:
		bit = win.SS_CENTER

	case AlignFar:
		bit = win.SS_RIGHT

	default:
		bit = win.SS_LEFT
	}

	return l.ensureStyleBits(bit, true)
}

func (l *Label) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_NCHITTEST:
		return win.HTCLIENT

	case win.WM_SETTEXT:
		l.textChangedPublisher.Publish()

	case win.WM_SIZE, win.WM_SIZING:
		l.Invalidate()
	}

	return l.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
