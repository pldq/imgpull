package display

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"

	"github.com/buger/goterm"
	"github.com/morikuni/aec"
)

// Full creates an EventProcessor that render advanced UI within a terminal.
// On Start, TUI lists task with a progress timer
func Full(out io.Writer) EventProcessor {
	return &ttyWriter{
		out:  out,
		done: make(chan bool),
		mtx:  &sync.Mutex{},
	}
}

type ttyWriter struct {
	out       io.Writer
	task      task
	done      chan bool
	mtx       *sync.Mutex
	ticker    *time.Ticker
	numLines  int
	lineWidth int
}

type task struct {
	ID        string
	startTime time.Time
	endTime   time.Time
	text      string
	details   string
	status    EventStatus
	current   int64
	percent   int
	total     int64
	spinner   *Spinner
}

func newTask(e Event) task {
	t := task{
		ID:        e.ID,
		startTime: time.Now(),
		text:      e.Text,
		details:   e.Details,
		status:    e.Status,
		current:   e.Current,
		percent:   e.Percent,
		total:     e.Total,
		spinner:   NewSpinner(),
	}
	if e.Status == Done || e.Status == Error {
		t.stop()
	}
	return t
}

// update adjusts task state based on last received event
func (t *task) update(e Event) {
	// update task based on received event
	switch e.Status {
	case Done, Error, Warning:
		if t.status != e.Status {
			t.stop()
		}
	case Working:
		t.hasMore()
	}
	t.status = e.Status
	t.text = e.Text
	t.details = e.Details
	// progress can only go up
	if e.Total == 0 || e.Total > t.total {
		t.total = e.Total
	}
	if e.Current == 0 || e.Current > t.current {
		t.current = e.Current
	}
	if e.Percent == 0 || e.Percent > t.percent {
		t.percent = e.Percent
	}
}

func (t *task) stop() {
	t.endTime = time.Now()
	t.spinner.Stop()
}

func (t *task) hasMore() {
	t.spinner.Restart()
}

func (t *task) Completed() bool {
	switch t.status {
	case Done, Error, Warning:
		return true
	default:
		return false
	}
}

func (w *ttyWriter) Start(ctx context.Context, event Event) {
	w.task = newTask(event)
	w.ticker = time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ctx.Done():
				// interrupted
				w.ticker.Stop()
				return
			case <-w.done:
				return
			case <-w.ticker.C:
				w.print()
			}
		}
	}()
}

func (w *ttyWriter) Done() {
	w.print()
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.ticker != nil {
		w.ticker.Stop()
	}
	w.done <- true
}

func (w *ttyWriter) On(event Event) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.event(event)
}

func (w *ttyWriter) OnText(event *Event, text string) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	event.Text = text
	w.event(*event)
}

func (w *ttyWriter) OnDetails(event *Event, details string) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	event.Details = details
	w.event(*event)
}

func (w *ttyWriter) event(e Event) {
	w.task.update(e)
}

// lineData holds pre-computed formatting for a task line
type lineData struct {
	spinner     string // rendered spinner with color
	prefix      string // dry-run prefix if any
	taskID      string // possibly abbreviated
	progress    string // progress bar and size info
	status      string // rendered status with color
	details     string // possibly abbreviated
	timer       string // rendered timer with color
	statusPad   int    // padding before status to align
	timerPad    int    // padding before timer to align
	statusColor colorFunc
}

func (w *ttyWriter) print() {
	terminalWidth := goterm.Width()
	terminalHeight := goterm.Height()
	if terminalWidth <= 0 {
		terminalWidth = 80
	}
	if terminalHeight <= 0 {
		terminalHeight = 24
	}
	w.printWithDimensions()
}

func (w *ttyWriter) printWithDimensions() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	up := w.numLines
	b := aec.NewBuilder(
		aec.Hide, // Hide the cursor while we are printing
		aec.Up(uint(up)),
		aec.Column(0),
	)
	_, _ = fmt.Fprint(w.out, b.ANSI)
	defer func() {
		_, _ = fmt.Fprint(w.out, aec.Show)
	}()

	line := w.prepareLineData(&w.task)
	w.applyPadding(&line)
	_, _ = fmt.Fprint(w.out, lineText(line))
	w.numLines = 1
}

func (w *ttyWriter) applyPadding(l *lineData) {
	// Width before statusPad: space(1) + spinner(1) + prefix + space(1) + taskID + progress
	beforeStatus := 2 +
		//lenAnsi(l.prefix) +
		//utf8.RuneCountInString(l.taskID) +
		lenAnsi(l.progress)
	// statusPad aligns status; lineText adds 1 more space after statusPad

	// Format: beforeStatus + statusPad + space(1) + status
	lineLen := beforeStatus + 1 + runewidth.StringWidth(l.status)
	if l.details != "" {
		lineLen += 1 + runewidth.StringWidth(l.details)
	}
	l.statusPad = lineLen
	l.timerPad = max(w.lineWidth-lineLen+1, 0)
	w.lineWidth = max(w.lineWidth, lineLen)
}

func (w *ttyWriter) prepareLineData(t *task) lineData {
	endTime := time.Now()
	if t.status != Working {
		endTime = t.startTime
		if (t.endTime != time.Time{}) {
			endTime = t.endTime
		}
	}

	prefix := ""

	elapsed := endTime.Sub(t.startTime).Seconds()

	var (
		hideDetails bool
		total       = t.total
		current     = t.current
		completion  []string
	)

	// only show the aggregated progress while the root operation is in-progress
	if t.status == Working && t.total == 0 {
		hideDetails = true
	}
	r := len(percentChars) - 1
	p := min(t.percent, 100)
	completion = append(completion, percentChars[r*p/100])

	var progress string
	if total > 0 && len(completion) > 0 {
		progress = " [" + SuccessColor(strings.Join(completion, "")) + "]"
		if !hideDetails {
			progress += fmt.Sprintf(" %7f / %-7f", float64(current), float64(total))
		}
	}

	return lineData{
		spinner:     spinner(t),
		prefix:      prefix,
		taskID:      t.ID,
		progress:    progress,
		status:      t.text,
		statusColor: colorFn(t.status),
		details:     t.details,
		timer:       fmt.Sprintf("%.1fs", elapsed),
	}
}

func lineText(l lineData) string {

	var sb strings.Builder
	sb.WriteString(" ")
	sb.WriteString(l.spinner)
	sb.WriteString(" ")
	sb.WriteString(l.statusColor(l.status))
	if l.details != "" {
		sb.WriteString(" ")
		sb.WriteString(l.details)
	}
	sb.WriteString(l.progress)
	sb.WriteString(strings.Repeat(" ", l.timerPad))
	//sb.WriteString(TimerColor(l.timer))
	sb.WriteString("\n")
	return sb.String()
}

var (
	spinnerDone    = "✔"
	spinnerWarning = "!"
	spinnerError   = "✘"
)

func spinner(t *task) string {
	switch t.status {
	case Done:
		return SuccessColor(spinnerDone)
	case Warning:
		return WarningColor(spinnerWarning)
	case Error:
		return ErrorColor(spinnerError)
	default:
		return CountColor(t.spinner.String())
	}
}

func colorFn(s EventStatus) colorFunc {
	switch s {
	case Done:
		return SuccessColor
	case Warning:
		return WarningColor
	case Error:
		return ErrorColor
	default:
		return nocolor
	}
}

// lenAnsi count of user-perceived characters in ANSI string.
func lenAnsi(s string) int {
	length := 0
	ansiCode := false
	for _, r := range s {
		if r == '\x1b' {
			ansiCode = true
			continue
		}
		if ansiCode && r == 'm' {
			ansiCode = false
			continue
		}
		if !ansiCode {
			length++
		}
	}
	return length
}

var percentChars = strings.Split("⠀⡀⣀⣄⣤⣦⣶⣷⣿", "")
