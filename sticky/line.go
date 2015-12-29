package sticky

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type line struct {
	hidden  bool
	linenum int
	parts   []Part
	output  *bufio.Writer
}

type Line interface {
	nudgeUp()
	render()
	hide()
	Display(text string, params ...interface{}) TextPart
	DisplayP(parts ...Part)
}

func NewLine(display string) Line {
	l := &line{
		hidden:  false,
		linenum: 1,
		parts:   nil,
		output:  bufio.NewWriter(os.Stdout),
	}

	l.Display(display)

	return l
}

func (l *line) nudgeUp() {
	l.linenum += 1

	l.DisplayP(l.parts...)
}

func (l *line) hide() {
	l.hidden = true
}

func (l *line) Display(text string, params ...interface{}) TextPart {
	part := NewTextPart(text, params...)

	l.DisplayP(part)

	return part
}

func (l *line) DisplayP(parts ...Part) {
	l.parts = parts

	for _, part := range l.parts {
		part.register(l)
	}

	l.render()
}

func (l *line) writeString(out string) {
	_, err := l.output.WriteString(out)
	if err != nil {
		log.Fatalf("Write String error: %s\n", err)
	}
}

func (l *line) render() {
	if l.hidden {
		return
	}

	mutex.Lock()

	l.writeString(fmt.Sprintf("\033[%dA", l.linenum)) // Move up to correct line
	l.writeString("\033[K")                           // Clear line

	width := 0
	for _, part := range l.parts {
		text := part.Render()
		l.output.WriteString(text)
		width += len(text)
	}

	l.writeString(fmt.Sprintf("\033[%dD", width))     // Move beginning of line
	l.writeString(fmt.Sprintf("\033[%dB", l.linenum)) // Move down to default line

	err := l.output.Flush()
	if err != nil {
		log.Fatalf("Output flush error: %s\n", err)
	}

	mutex.Unlock()
}
