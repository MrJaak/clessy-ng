package utils

import (
	"strings"
	"unicode"

	"git.fromouter.space/crunchy-rocks/draw2d"
)

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func SplitCenter(text string) []string {
	centerIndex := int(len(text) / 2)
	whitespaceFrontIndex := strings.IndexRune(text[centerIndex:], ' ')
	whitespaceBackIndex := strings.LastIndex(text[:centerIndex], " ")
	frontIndex := centerIndex + whitespaceFrontIndex
	frontDiff := 9999
	backDiff := 9999
	if whitespaceFrontIndex < 1 && whitespaceBackIndex < 1 {
		return []string{text}
	}
	if whitespaceFrontIndex >= 0 {
		frontDiff = len(text[:frontIndex]) - len(text[frontIndex:])
	}
	if whitespaceBackIndex >= 0 {
		backDiff = len(text[:whitespaceBackIndex]) - len(text[whitespaceBackIndex:])
	}
	if abs(frontDiff) < abs(backDiff) {
		return []string{strings.TrimSpace(text[:frontIndex]), strings.TrimSpace(text[frontIndex:])}
	}
	return []string{strings.TrimSpace(text[:whitespaceBackIndex]), strings.TrimSpace(text[whitespaceBackIndex:])}
}

// Word wrapping code from https://github.com/fogleman/gg
// Copyright (C) 2016 Michael Fogleman
// Licensed under MIT (https://github.com/fogleman/gg/blob/master/LICENSE.md)

func splitOnSpace(x string) []string {
	var result []string
	pi := 0
	ps := false
	for i, c := range x {
		s := unicode.IsSpace(c)
		if s != ps && i > 0 {
			result = append(result, x[pi:i])
			pi = i
		}
		ps = s
	}
	result = append(result, x[pi:])
	return result
}

func WordWrap(gc draw2d.GraphicContext, s string, width float64) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		fields := splitOnSpace(line)
		if len(fields)%2 == 1 {
			fields = append(fields, "")
		}
		x := ""
		for i := 0; i < len(fields); i += 2 {
			left, _, right, _ := gc.GetStringBounds(x + fields[i])
			w := right - left
			if w > width {
				if x == "" {
					result = append(result, fields[i])
					x = ""
					continue
				} else {
					result = append(result, x)
					x = ""
				}
			}
			x += fields[i] + fields[i+1]
		}
		if x != "" {
			result = append(result, x)
		}
	}
	for i, line := range result {
		result[i] = strings.TrimSpace(line)
	}
	return result
}
