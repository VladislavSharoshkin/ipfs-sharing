package elements

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type ToolbarLabel struct {
	*widget.Label
}

func NewToolbarLabel(label string) *ToolbarLabel {
	L := widget.NewLabelWithStyle(label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	L.MinSize()
	return &ToolbarLabel{L}
}

func (t *ToolbarLabel) ToolbarObject() fyne.CanvasObject {
	return t.Label
}
