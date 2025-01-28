package main

import (
	lg "github.com/charmbracelet/lipgloss"
	"github.com/mkiene/huh"
)

var tree_style = lg.NewStyle().Border(lg.NormalBorder()).Padding(0, 1, 0, 1)
var bold_style = lg.NewStyle().Bold(true).Foreground(lg.Color("#ffffff"))

var form_theme = huh.ThemeBase()
