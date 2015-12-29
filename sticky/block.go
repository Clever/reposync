package sticky

import (
	"bufio"
	"log"
	"os"
	"sync"
)

type Block struct {
	hidden bool
	output *bufio.Writer
	lines  []Line
	Height int
}

func NewBlock(height int) *Block {
	block := &Block{
		hidden: false,
		output: bufio.NewWriter(os.Stdout),
		Height: height,
	}

	block.AddLines(height)

	return block
}

var mutex = &sync.Mutex{}

func (b *Block) AddLines(num int) {
	for i := 0; i < num; i++ {
		b.nudgeUp()

		line := NewLine("")
		b.lines = append(b.lines, line)

		if b.hidden {
			line.hide()
		}
	}
}

func (b *Block) Line(idx int) Line {
	if idx < 0 {
		return b.lines[len(b.lines)+idx]
	} else {
		return b.lines[idx]
	}
}

func (b *Block) Hide() {
	b.hidden = true

	for _, line := range b.lines {
		line.hide()
	}
}

func (b *Block) nudgeUp() {
	if !b.hidden {
		mutex.Lock()

		_, err := b.output.WriteString("\n")
		if err != nil {
			log.Fatalf("Nudge output error: %s\n", err)
		}

		err = b.output.Flush()
		if err != nil {
			log.Fatalf("Nudge flush output error: %s\n", err)
		}

		mutex.Unlock()
	}

	for _, line := range b.lines {
		line.nudgeUp()
	}
}
