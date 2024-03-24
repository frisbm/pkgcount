package models

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"text/template"
)

// PackageCount contains a package and its count.
type PackageCount struct {
	Package string
	Count   int
}

// Result contains the internal and external package counts.
type Result struct {
	Internal []PackageCount
	External []PackageCount
}

const (
	markdownTemplate = `
**Internal Package Counts**

| Package        |        Count |
| :---           |         ---: |
{{- if isEmpty .Internal }}
|       -        |      0       |
{{- else }}
{{- range .Internal }}
| {{- .Package }} | {{- .Count }} |
{{- end }}
{{- end }}

**External Package Counts**

| Package        |        Count |
| :---           |         ---: |
{{- if isEmpty .External }}
|       -        |      0       |
{{- else }}
{{- range .External }}
| {{- .Package }} | {{- .Count }} |
{{- end }}
{{- end }}
`
)

var (
	funcMap = map[string]any{
		"isEmpty": func(items []PackageCount) bool {
			return len(items) == 0
		},
	}
)

// Unrendered generates markdown output with the internal and external package counts.
func (r Result) Unrendered() ([]byte, error) {
	tmpl, err := template.New("markdown.tmpl").Funcs(funcMap).Parse(strings.TrimLeft(markdownTemplate, "\n"))
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if err = tmpl.Execute(&b, r); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Rendered generates markdown output with the internal and external package counts.
func (r Result) Rendered() ([]byte, error) {
	if len(r.Internal) == 0 {
		r.Internal = []PackageCount{{Package: "-", Count: 0}}
	}
	if len(r.External) == 0 {
		r.External = []PackageCount{{Package: "-", Count: 0}}
	}

	var buffer bytes.Buffer
	renderPackageSet("Internal", r.Internal, &buffer)
	buffer.WriteString("\n")
	renderPackageSet("External", r.External, &buffer)

	return buffer.Bytes(), nil
}

const (
	space                 = " "
	leftTopCorner         = "┌"
	leftBottomCorner      = "└"
	rightTopCorner        = "┐"
	rightBottomCorner     = "┘"
	horizontalLine        = "─"
	doubleHorizontal      = "═"
	verticalLine          = "│"
	topTee                = "┬"
	bottomTee             = "┴"
	leftTee               = "├"
	rightTee              = "┤"
	cross                 = "┼"
	doubleLeftTee         = "╞"
	doubleRightTee        = "╡"
	doubleHorizontalCross = "╪"
)

func renderPackageSet(packageType string, packagecounts []PackageCount, buffer *bytes.Buffer) {
	widest := slices.MaxFunc(packagecounts, widestPackageNameFunc)
	width := max(len(widest.Package), 8)

	// ┌───────────┬─────┐
	buffer.WriteString(fmt.Sprintf("\033[1m%s Package Counts\033[0m\n", packageType))
	buffer.WriteString(leftTopCorner)
	buffer.WriteString(strings.Repeat(horizontalLine, width+2))
	buffer.WriteString(topTee)
	buffer.WriteString(strings.Repeat(horizontalLine, 5))
	buffer.WriteString(rightTopCorner)
	buffer.WriteString("\n")

	// │Package    │Count│
	buffer.WriteString(verticalLine)
	buffer.WriteString("Package")
	buffer.WriteString(strings.Repeat(space, width-5))
	buffer.WriteString(verticalLine)
	buffer.WriteString("Count")
	buffer.WriteString(verticalLine)
	buffer.WriteString("\n")

	// ╞═══════════╪═════╡
	buffer.WriteString(doubleLeftTee)
	buffer.WriteString(strings.Repeat(doubleHorizontal, width+2))
	buffer.WriteString(doubleHorizontalCross)
	buffer.WriteString(strings.Repeat(doubleHorizontal, 5))
	buffer.WriteString(doubleRightTee)
	buffer.WriteString("\n")

	// │    ...    │  ...│
	for i, pc := range packagecounts {
		buffer.WriteString(verticalLine)
		buffer.WriteString(fmt.Sprintf(" %-*s ", width, pc.Package))
		buffer.WriteString(verticalLine)
		buffer.WriteString(fmt.Sprintf(" %4d", pc.Count))
		buffer.WriteString(verticalLine)
		buffer.WriteString("\n")

		if i == len(packagecounts)-1 {
			break
		}

		// ├───────────┼─────┤
		buffer.WriteString(leftTee)
		buffer.WriteString(strings.Repeat(horizontalLine, width+2))
		buffer.WriteString(cross)
		buffer.WriteString(strings.Repeat(horizontalLine, 5))
		buffer.WriteString(rightTee)
		buffer.WriteString("\n")
	}
	// └───────────┴─────┘
	buffer.WriteString(leftBottomCorner)
	buffer.WriteString(strings.Repeat(horizontalLine, width+2))
	buffer.WriteString(bottomTee)
	buffer.WriteString(strings.Repeat(horizontalLine, 5))
	buffer.WriteString(rightBottomCorner)
	buffer.WriteString("\n")
}

func widestPackageNameFunc(i, j PackageCount) int {
	return len(i.Package) - len(j.Package)
}
