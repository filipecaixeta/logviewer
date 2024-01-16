package logs

import (
	"strings"

	"github.com/filipecaixeta/logviewer/internal/logs/pipeline"
)

type circularLogBuffer struct {
	Buffer []pipeline.LogEntry
	Head   int
	Tail   int
}

// Adds a new element and returns the one removed
func (c *circularLogBuffer) Add(t pipeline.LogEntry) (removed *pipeline.LogEntry) {
	if len(c.Buffer) == cap(c.Buffer) {
		// Buffer is full
		removed = &c.Buffer[c.Tail]
		c.Buffer[c.Tail] = t
		c.Tail = (c.Tail + 1) % cap(c.Buffer)
		if c.Tail == c.Head {
			c.Head = (c.Head + 1) % cap(c.Buffer)
		}
	} else {
		// Buffer is not full
		c.Buffer = append(c.Buffer, t)
		c.Tail = len(c.Buffer) % cap(c.Buffer)
	}
	if c.Tail == c.Head && len(c.Buffer) > 0 {
		c.Head = (c.Head + 1) % cap(c.Buffer)
	}
	return removed
}

func (c *circularLogBuffer) First() *pipeline.LogEntry {
	if len(c.Buffer) == 0 {
		return nil
	}
	return &c.Buffer[c.Head]
}

func (c *circularLogBuffer) Last() *pipeline.LogEntry {
	if len(c.Buffer) == 0 {
		return nil
	}
	lastIndex := c.Tail - 1
	if lastIndex < 0 {
		lastIndex = cap(c.Buffer) - 1
	}
	return &c.Buffer[lastIndex]
}

func (c *circularLogBuffer) Height() int {
	if len(c.Buffer) == 0 {
		return 0
	}
	lastHeight := c.Last().CumHeight
	return lastHeight
}

func (c *circularLogBuffer) RunPipeline(f func(l *pipeline.LogEntry) error) {
	for i := c.Head; i != c.Tail; i = (i + 1) % cap(c.Buffer) {
		_ = f(&c.Buffer[i])
	}
}

// iterate over the elements in the buffer based on the current scroll offset
// returns the lines that are visible
func (c *circularLogBuffer) View(scroll int, height int) string {
	if len(c.Buffer) == 0 {
		return ""
	}

	if l := c.Last(); scroll > l.CumHeight-l.Height {
		scroll = l.CumHeight - l.Height
	}

	if f := c.First(); scroll < f.CumHeight-f.Height {
		scroll = f.CumHeight - f.Height
	}

	firstVisible := c.binarySearchFirstVisible(scroll)

	var lineCount int
	var b strings.Builder

	// Iterate over the elements in the buffer until the end of the screen is reached
	for i := firstVisible; ; i = (i + 1) % cap(c.Buffer) {
		if i == c.Tail || lineCount == height {
			break
		}

		if !c.Buffer[i].Show {
			continue
		}

		// height of the first line of the log
		firstLineOffset := c.Buffer[i].CumHeight - c.Buffer[i].Height

		// handle cases where the first line of the log is not fully visible
		// or there is only one line in the log and it's height is greater than the height of the screen
		if i == firstVisible && (scroll > firstLineOffset || c.Buffer[i].Height > height) {
			formatted := c.Buffer[i].Formatted
			n := scroll - firstLineOffset
			lineHeight := c.Buffer[i].Height - n
			p := findLinePos(formatted, n)
			formatted = formatted[p:]
			if lineHeight > height {
				lineHeight = height
				formatted = formatted[:findLinePos(formatted, height)]
			}
			b.WriteString(formatted)
			lineCount += lineHeight
			if lineCount < height {
				b.WriteString("\n")
			}
			continue
		}

		// if the line is too long to fit on the screen, find the position of the last \n that fits on the screen
		if lineCount+c.Buffer[i].Height > height {
			formatted := c.Buffer[i].Formatted
			// find the position of the last \n that fits on the screen
			lineHeight := height - lineCount
			p := findLinePos(formatted, lineHeight)
			b.WriteString(formatted[:p])
			lineCount += lineHeight
			if lineCount < height {
				b.WriteString("\n")
			}
			break
		}
		b.WriteString(c.Buffer[i].Formatted)
		b.WriteString("\n")
		lineCount += c.Buffer[i].Height
	}

	// fill with \n until the end of the screen is reached
	for i := 0; i < height-lineCount; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

func (c *circularLogBuffer) GetLogEntryAtScrollOffset(scroll int) *pipeline.LogEntry {
	if len(c.Buffer) == 0 {
		return nil
	}

	logEntryIndex := c.binarySearchFirstVisible(scroll)
	if logEntryIndex < 0 {
		return nil
	}
	return &c.Buffer[logEntryIndex]
}

func (c *circularLogBuffer) binarySearchFirstVisible(scroll int) int {
	if l := c.Last(); l != nil && scroll > l.CumHeight {
		return -1
	}

	start := c.Head
	end := c.Tail
	if start == end {
		return start
	}

	var firstVisible int
	for start != end {
		var mid int
		if start < end {
			mid = (start + (end-start)/2) % cap(c.Buffer)
		} else {
			mid = (start + (cap(c.Buffer)-start+end)/2) % cap(c.Buffer)
		}
		midEntry := c.Buffer[mid]
		midCumHeight := midEntry.CumHeight
		midStartHeight := midCumHeight - midEntry.Height

		if midStartHeight <= scroll && midCumHeight > scroll {
			firstVisible = mid
			break
		}

		if midCumHeight <= scroll {
			start = (mid + 1) % cap(c.Buffer)
		} else {
			end = mid
		}
	}

	return firstVisible
}

// findLinePos finds the position of the first caracter after the n-th line
func findLinePos(s string, n int) int {
	if n == 0 {
		return 0
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			n--
			if n == 0 {
				return i + 1
			}
		}
	}
	// the last line doesn't have a \n
	// so it's handled as a special case
	if n == 1 {
		return len(s)
	}
	return -1
}
