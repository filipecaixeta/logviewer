package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	jsoniter "github.com/json-iterator/go"

	"github.com/filipecaixeta/logviewer/internal/config"

	"github.com/alecthomas/chroma/quick"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/mattn/go-runewidth"
	"github.com/mitchellh/go-wordwrap"
)

var jsonLib = jsoniter.ConfigCompatibleWithStandardLibrary

type LogEntry struct {
	Show      bool
	Raw       string
	Formatted string
	Json      map[string]interface{}
	Height    int
	CumHeight int
	Index     int
}

func numberToGoTypes(j interface{}) interface{} {
	switch v := j.(type) {
	case map[string]interface{}:
		for key, value := range v {
			v[key] = numberToGoTypes(value)
		}
	case []interface{}:
		for i, value := range v {
			v[i] = numberToGoTypes(value)
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			j = i
		}
		if f, err := v.Float64(); err == nil {
			j = f
		}
	}
	return j
}

func runToJson(l *LogEntry) error {
	l.Json = nil
	if l.Raw[0] == '{' {
		d := jsonLib.NewDecoder(strings.NewReader(l.Raw))
		d.UseNumber()
		if err := d.Decode(&l.Json); err != nil {
			l.Json = nil
			return nil
		}
		numberToGoTypes(l.Json)
	}
	return nil
}

type LogFormat struct {
	ReturnedFields []string
	Width          uint
	Highlight      bool
}

func (lt *LogFormat) RunReturnedFieldsAndFormat(l *LogEntry) error {
	if !l.Show {
		l.Height = 0
		l.Formatted = ""
		return nil
	}
	if len(l.Json) == 0 {
		l.Formatted = l.Raw
		if lt.Width != 0 {
			l.Formatted = wordwrap.WrapString(l.Formatted, lt.Width)
		}
		l.Height = len(strings.Split(l.Formatted, "\n"))
		return nil
	}

	j := l.Json

	if len(lt.ReturnedFields) != 0 {
		j = make(map[string]interface{})
		for _, field := range lt.ReturnedFields {
			switch {
			case strings.HasPrefix(field, "*") && strings.HasSuffix(field, "*"):
				// Handling prefix and suffix wildcards
				middle := strings.Trim(field, "*")
				for key, value := range l.Json {
					if strings.Contains(key, middle) {
						j[key] = value
					}
				}

			case strings.HasPrefix(field, "*"):
				// Handling suffix wildcards
				suffix := strings.TrimPrefix(field, "*")
				for key, value := range l.Json {
					if strings.HasSuffix(key, suffix) {
						j[key] = value
					}
				}

			case strings.HasSuffix(field, "*"):
				// Handling prefix wildcards
				prefix := strings.TrimSuffix(field, "*")
				for key, value := range l.Json {
					if strings.HasPrefix(key, prefix) {
						j[key] = value
					}
				}

			default:
				// No wildcards
				fieldParts := strings.Split(field, ".")
				addToResult(l.Json, fieldParts, 0, j)
			}
		}
	}

	jsonLog, _ := jsonLib.MarshalIndent(j, "", " ")
	if lt.Highlight {
		var buff bytes.Buffer
		_ = quick.Highlight(&buff, string(jsonLog), "json", "terminal256", config.Theme)
		l.Formatted = buff.String()
	} else {
		l.Formatted = string(jsonLog)
	}

	if lt.Width != 0 {
		l.Formatted = WrapString(l.Formatted, int(lt.Width)-5)
	}
	l.Height = len(strings.Split(l.Formatted, "\n"))

	return nil
}

// addToResult recursively navigates through the Json map and adds the specified field to the result.
func addToResult(currentMap map[string]interface{}, fieldParts []string, index int, result map[string]interface{}) {
	if index == len(fieldParts)-1 {
		if value, exists := currentMap[fieldParts[index]]; exists {
			// Add the final value to the result
			insertIntoMap(result, fieldParts, value)
		}
		return
	}

	if nextMap, ok := currentMap[fieldParts[index]].(map[string]interface{}); ok {
		// Create a new map if it doesn't exist at the current level in result
		if _, exists := result[fieldParts[index]]; !exists {
			result[fieldParts[index]] = make(map[string]interface{})
		}
		// Continue to the next level
		addToResult(nextMap, fieldParts, index+1, result[fieldParts[index]].(map[string]interface{}))
	}
}

// insertIntoMap inserts the value into the result map, maintaining the structure defined by fieldParts.
func insertIntoMap(result map[string]interface{}, fieldParts []string, value interface{}) {
	for i := len(fieldParts) - 1; i >= 0; i-- {
		if i == len(fieldParts)-1 {
			result[fieldParts[i]] = value
		} else {
			result = map[string]interface{}{fieldParts[i]: result}
		}
	}
}

type LogFilter struct {
	FilterExpr string
	Filter     *vm.Program
	Default    bool
}

func level2Int(level string) int {
	switch strings.ToLower(level) {
	case "debug":
		return 0
	case "info":
		return 1
	case "warn", "warning":
		return 2
	case "error":
		return 3
	case "fatal", "panic":
		return 4
	}
	return 0
}

// filter a log entry based on the level
// filterLevel(logLevel, minLevel)
var filterLevel = expr.Function(
	"filterLevel",
	func(params ...any) (any, error) {
		logLevel := level2Int(params[0].(string))
		minLevel := level2Int(params[1].(string))
		return logLevel >= minLevel, nil
	},
	new(func(string, string) bool),
)

var toLocalDateStr = expr.Function(
	"toLocalDateStr",
	func(params ...any) (any, error) {
		ts := params[0].(float64)
		return time.Unix(int64(ts), 0).In(time.Local).Format("2006-01-02 15:04:05"), nil
	},
	new(func(float64) string),
)

var toDateStr = expr.Function(
	"toDateStr",
	func(params ...any) (any, error) {
		ts := params[0].(float64)
		return time.Unix(int64(ts), 0).Format("2006-01-02 15:04:05"), nil
	},
	new(func(float64) string),
)

func (lf *LogFilter) Compile() error {
	if lf.FilterExpr == "" {
		lf.Filter = nil
		return nil
	}
	p, err := expr.Compile(lf.FilterExpr, filterLevel, expr.AsBool())
	if err != nil {
		lf.Filter = nil
		return err
	}
	lf.Filter = p
	return nil
}

func (lf *LogFilter) RunFilter(l *LogEntry) error {
	if lf.Filter == nil {
		l.Show = true
		return nil
	}
	r, err := vm.Run(lf.Filter, map[string]interface{}{
		"text": l.Raw,
		"json": l.Json,
	})
	if err != nil {
		l.Show = lf.Default
		return err
	} else if r == nil || r.(bool) {
		l.Show = true
	} else {
		l.Show = false
	}
	return nil
}

type LogTransform struct {
	Transforms []func(*LogEntry) error
}

func compileLogTransforms(transforms []config.Transform) []func(*LogEntry) error {
	tf := make([]func(*LogEntry) error, 0, len(transforms))
	for _, t := range transforms {
		p, err := expr.Compile(t.Expression, toDateStr, toLocalDateStr)
		if err != nil {
			fmt.Printf("error compiling expression: %v\n", err)
			continue
		}
		f := func(l *LogEntry) error {
			r, err := vm.Run(p, map[string]interface{}{
				"text": l.Raw,
				"json": l.Json,
			})
			if err != nil {
				fmt.Printf("error running expression: %v\n", err)
				return err
			}
			l.Json[t.Field] = r
			return nil
		}
		tf = append(tf, f)
	}
	return tf
}

func (lt *LogTransform) RunTransform(l *LogEntry) error {
	for _, f := range lt.Transforms {
		_ = f(l)
	}
	return nil
}

type LogPipeline struct {
	Pipeline  []func(*LogEntry) error
	index     int
	cumHeight int
	lf        *LogFilter
	lft       *LogFormat
	lt        *LogTransform
	Cfg       *config.View
}

func New(cfg *config.View, width uint) (*LogPipeline, error) {
	if cfg == nil {
		cfg = &config.View{}
	}
	lf := &LogFilter{
		Default: cfg.FilterDefault,
	}
	lf.FilterExpr = cfg.Filter
	if err := lf.Compile(); err != nil {
		return nil, err
	}
	lft := &LogFormat{
		ReturnedFields: cfg.ReturnedFields,
		Width:          width,
		Highlight:      true,
	}
	lt := &LogTransform{Transforms: compileLogTransforms(cfg.Transforms)}
	lp := &LogPipeline{
		lf:  lf,
		lft: lft,
		lt:  lt,
		Cfg: cfg,
	}
	lp.Pipeline = []func(*LogEntry) error{
		lp.setIndex,
		runToJson,
		lt.RunTransform,
		lf.RunFilter,
		lft.RunReturnedFieldsAndFormat,
		lp.setCumHeight,
	}
	return lp, nil
}

func (lp *LogPipeline) setIndex(l *LogEntry) error {
	lp.index++
	l.Index = lp.index
	return nil
}

func (lp *LogPipeline) setCumHeight(l *LogEntry) error {
	if l.Show {
		l.CumHeight = lp.cumHeight + l.Height
		lp.cumHeight = l.CumHeight
	} else {
		l.CumHeight = lp.cumHeight
	}
	return nil
}

func (lp *LogPipeline) Run(l *LogEntry) error {
	for _, f := range lp.Pipeline {
		_ = f(l)
	}
	return nil
}

func (lp *LogPipeline) Reset() {
	lp.cumHeight = 0
}

func (lp *LogPipeline) SetWidth(width uint) error {
	lp.Reset()
	lp.lft.Width = width
	return nil
}

func (lp *LogPipeline) RunWidthChanged(l *LogEntry) error {
	for _, f := range lp.Pipeline[4:] {
		_ = f(l)
	}
	return nil
}

func (lp *LogPipeline) SetFilter(filter string) error {
	lp.Reset()
	lp.lf.FilterExpr = filter
	if filter == "" {
		lp.lf.Filter = nil
		return nil
	}
	lp.Cfg.Filter = filter
	return lp.lf.Compile()
}

func (lp *LogPipeline) RunFilterChanged(l *LogEntry) error {
	for _, f := range lp.Pipeline[3:] {
		_ = f(l)
	}
	return nil
}

func (lp *LogPipeline) SetReturnedFields(fields []string) error {
	lp.Reset()
	for i := 0; i < len(fields); i++ {
		fields[i] = strings.TrimSpace(fields[i])
		if fields[i] == "" {
			fields = append(fields[:i], fields[i+1:]...)
			i--
		}
	}
	lp.lft.ReturnedFields = fields
	lp.Cfg.ReturnedFields = fields
	return nil
}

func (lp *LogPipeline) RunReturnedFieldsChanged(l *LogEntry) error {
	for _, f := range lp.Pipeline[4:] {
		_ = f(l)
	}
	return nil
}

func (lp *LogPipeline) SetTransforms(transforms []config.Transform) error {
	lp.Reset()
	lp.lt.Transforms = compileLogTransforms(transforms)
	lp.Cfg.Transforms = transforms
	return nil
}

func (lp *LogPipeline) SetView(view *config.View) error {
	lp.Reset()
	lp.Cfg = view
	lp.lf.Default = view.FilterDefault
	if err := lp.SetTransforms(view.Transforms); err != nil {
		return err
	}
	if err := lp.SetFilter(view.Filter); err != nil {
		return err
	}
	if err := lp.SetReturnedFields(view.ReturnedFields); err != nil {
		return err
	}
	return nil
}

func (lp *LogPipeline) RunViewChanged(l *LogEntry) error {
	for _, f := range lp.Pipeline[1:] {
		_ = f(l)
	}
	return nil
}

func WrapString(str string, width int) string {
	if width <= 0 {
		return str
	}

	var result strings.Builder
	var currentLine strings.Builder
	var currentLineWidth int
	var wordBuffer strings.Builder
	var inEscapeSequence bool
	var wordWidth int
	var escapeSequenceBuffer strings.Builder
	var escapeSequence string

	for _, runeValue := range str {
		// Handle ANSI escape sequences.
		if runeValue == '\x1b' {
			inEscapeSequence = true
		}
		if inEscapeSequence {
			escapeSequenceBuffer.WriteRune(runeValue)
			wordBuffer.WriteRune(runeValue)
			if ('a' <= runeValue && runeValue <= 'z') || ('A' <= runeValue && runeValue <= 'Z') {
				inEscapeSequence = false
			}
			continue
		}

		if e := escapeSequenceBuffer.String(); e != "" {
			escapeSequence = e
			escapeSequenceBuffer.Reset()
		}

		// Handle newline characters.
		if runeValue == '\n' {
			currentLine.WriteString(wordBuffer.String())
			result.WriteString(currentLine.String() + "\n")
			currentLine.Reset()
			currentLine.WriteString(escapeSequence)
			wordBuffer.Reset()
			currentLineWidth = 0
			wordWidth = 0
			continue
		}

		// Add rune to the word buffer.
		wordBuffer.WriteRune(runeValue)
		wordWidth += runewidth.RuneWidth(runeValue)

		// Check if the rune is a space or a wide character.
		if unicode.IsSpace(runeValue) || runewidth.RuneWidth(runeValue) > 1 {
			if currentLineWidth+wordWidth > width {
				// Wrap line at the width.
				result.WriteString(currentLine.String() + "\n")
				currentLine.Reset()
				currentLine.WriteString(escapeSequence)
				currentLineWidth = 0
			}
			currentLine.WriteString(wordBuffer.String())
			currentLineWidth += wordWidth
			wordBuffer.Reset()
			wordWidth = 0
		}
	}

	// Append any remaining text.
	if wordBuffer.Len() > 0 {
		if currentLineWidth+wordWidth > width {
			result.WriteString(currentLine.String() + "\n")
			currentLine.Reset()
			currentLine.WriteString(escapeSequence)
		}
		currentLine.WriteString(wordBuffer.String())
	}
	if currentLine.Len() > 0 {
		result.WriteString(currentLine.String())
	}

	return result.String()
}
