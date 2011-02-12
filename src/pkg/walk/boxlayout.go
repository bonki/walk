// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package walk

import (
	//"log"
	"os"
)

type Orientation byte

const (
	Horizontal Orientation = iota
	Vertical
)

type BoxLayout struct {
	container   Container
	margins     Margins
	spacing     int
	orientation Orientation
}

func NewHBoxLayout() *BoxLayout {
	return &BoxLayout{orientation: Horizontal}
}

func NewVBoxLayout() *BoxLayout {
	return &BoxLayout{orientation: Vertical}
}

func (l *BoxLayout) Container() Container {
	return l.container
}

func (l *BoxLayout) SetContainer(value Container) {
	if value != l.container {
		if l.container != nil {
			l.container.SetLayout(nil)
		}

		l.container = value

		if value != nil && value.Layout() != Layout(l) {
			value.SetLayout(l)

			l.Update(true)
		}
	}
}

func (l *BoxLayout) Margins() Margins {
	return l.margins
}

func (l *BoxLayout) SetMargins(value Margins) os.Error {
	if value.HNear < 0 || value.VNear < 0 || value.HFar < 0 || value.VFar < 0 {
		return newError("margins must be positive")
	}

	l.margins = value

	return nil
}

func (l *BoxLayout) Orientation() Orientation {
	return l.orientation
}

func (l *BoxLayout) SetOrientation(value Orientation) os.Error {
	if value != l.orientation {
		switch value {
		case Horizontal, Vertical:

		default:
			return newError("invalid Orientation value")
		}

		l.orientation = value

		l.Update(false)
	}

	return nil
}

func (l *BoxLayout) Spacing() int {
	return l.spacing
}

func (l *BoxLayout) SetSpacing(value int) os.Error {
	if value != l.spacing {
		if value < 0 {
			return newError("spacing cannot be negative")
		}

		l.spacing = value

		l.Update(false)
	}

	return nil
}

func (l *BoxLayout) Update(reset bool) (err os.Error) {
	if l.container == nil {
		return
	}

	/*log.Printf("*BoxLayout.Update: HWND: %d\n", l.container.Handle())
	for i, child := range(l.container.Children().items) {
		log.Printf("children: index: %d, type: %T\n", i, child)
	}*/

	widgets := make([]Widget, 0, l.container.Children().Len())

	children := l.container.Children()
	j := 0
	for i := 0; i < cap(widgets); i++ {
		widget := children.At(i)

		ps := widget.PreferredSize()
		if ps.Width == 0 && ps.Height == 0 && widget.LayoutFlags()&widget.LayoutFlagsMask() == 0 {
			continue
		}

		widgets = widgets[0 : j+1]
		widgets[j] = widget
		j++
	}

	widgetCount := len(widgets)

	if widgetCount == 0 {
		return
	}

	// We will start by collecting some valuable information.
	flags := make([]LayoutFlags, widgetCount)
	prefSizes := make([]Size, widgetCount)
	var prefSizeSum Size
	var shrinkHorzCount, growHorzCount, shrinkVertCount, growVertCount int

	for i := 0; i < widgetCount; i++ {
		widget := widgets[i]

		ps := widget.PreferredSize()

		maxSize := widget.MaxSize()

		lf := widget.LayoutFlags() & widget.LayoutFlagsMask()
		if maxSize.Width > 0 {
			lf &^= HGrow
			ps.Width = maxSize.Width
		}
		if maxSize.Height > 0 {
			lf &^= VGrow
			ps.Height = maxSize.Height
		}

		if lf&HShrink > 0 {
			shrinkHorzCount++
		}
		if lf&HGrow > 0 {
			growHorzCount++
		}
		if lf&VShrink > 0 {
			shrinkVertCount++
		}
		if lf&VGrow > 0 {
			growVertCount++
		}
		flags[i] = lf

		prefSizeSum.Width += ps.Width
		prefSizeSum.Height += ps.Height
		prefSizes[i] = ps
	}

	cb := l.container.ClientBounds()

	spacingSum := (widgetCount - 1) * l.spacing

	// Now do the actual layout thing.
	if l.orientation == Vertical {
		diff := cb.Height - l.margins.VNear - prefSizeSum.Height - spacingSum - l.margins.VFar

		reqW := 0

		for i, s := range prefSizes {
			if s.Width > reqW && (flags[i]&HShrink == 0) {
				reqW = s.Width
			}
		}
		//        if reqW == 0 {
		reqW = cb.Width - l.margins.HNear - l.margins.HFar
		//        }

		var change int
		if diff < 0 {
			if shrinkVertCount > 0 {
				change = diff / shrinkVertCount
			}
		} else {
			if growVertCount > 0 {
				change = diff / growVertCount
			}
		}

		//log.Printf("*BoxLayout.Update: widgetCount: %d, cb: %+v, prefSizeSum: %+v, diff: %d, change: %d, reqW: %d", widgetCount, cb, prefSizeSum, diff, change, reqW)

		y := cb.Y + l.margins.VNear
		for i := 0; i < widgetCount; i++ {
			widget := widgets[i]

			h := prefSizes[i].Height

			switch {
			case change < 0:
				if flags[i]&VShrink > 0 {
					h += change
				}

			case change > 0:
				if flags[i]&VGrow > 0 {
					h += change
				}
			}

			bounds := Rectangle{cb.X + l.margins.HNear, y, reqW, h}

			//log.Printf("*BoxLayout.Update: bounds: %+v", bounds)

			widget.SetBounds(bounds)

			y += h + l.spacing
		}
	} else {
		diff := cb.Width - l.margins.HNear - prefSizeSum.Width - spacingSum - l.margins.HFar
		reqH := 0

		for i, s := range prefSizes {
			if s.Height > reqH && (flags[i]&VShrink == 0) {
				reqH = s.Height
			}
		}
		//        if reqH == 0 {
		reqH = cb.Height - l.margins.VNear - l.margins.VFar
		//        }

		var change int
		if diff < 0 {
			if shrinkHorzCount > 0 {
				change = diff / shrinkHorzCount
			}
		} else {
			if growHorzCount > 0 {
				change = diff / growHorzCount
			}
		}

		//log.Printf("*BoxLayout.Update: widgetCount: %d, cb: %+v, prefSizeSum: %+v, diff: %d, change: %d, reqH: %d", widgetCount, cb, prefSizeSum, diff, change, reqH)

		x := cb.X + l.margins.HNear
		for i := 0; i < widgetCount; i++ {
			widget := widgets[i]

			w := prefSizes[i].Width

			switch {
			case change < 0:
				if flags[i]&HShrink > 0 {
					w += change
				}

			case change > 0:
				if flags[i]&HGrow > 0 {
					w += change
				}
			}

			bounds := Rectangle{x, cb.Y + l.margins.VNear, w, reqH}

			//log.Printf("*BoxLayout.Update: bounds: %+v", bounds)

			widget.SetBounds(bounds)

			x += w + l.spacing
		}
	}

	return
}