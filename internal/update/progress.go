package update

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Explicit installer/update commands alone attach this display. Background
// checks and library callers remain silent; stdout is always the JSON result.
type progressDisplay struct {
	out  io.Writer
	last time.Time
}

func progressEnabled() bool {
	return terminal(os.Stderr.Fd()) && os.Getenv("TERM") != "dumb" && os.Getenv("MOI_AGENT_CALL") != "1" && os.Getenv("CI") == ""
}

func (p *progressDisplay) clear() {
	if p != nil {
		fmt.Fprint(p.out, "\r\x1b[2K")
	}
}

func (p *progressDisplay) stage(label string) {
	if p == nil {
		return
	}
	p.last = time.Time{}
	p.draw(label)
}

func (p *progressDisplay) draw(label string) {
	// All labels here are fixed ASCII, never URLs, paths, or server responses.
	var size struct{ rows, cols, xpixel, ypixel uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stderr.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&size)))
	if errno == 0 && size.cols > 0 {
		label = label[:min(len(label), int(size.cols)-1)]
	}
	fmt.Fprintf(p.out, "\r\x1b[2K%s", label)
}

func (p *progressDisplay) writer(out io.Writer, name string, total int64) io.Writer {
	if p == nil {
		return out
	}
	return &progressWriter{out: out, display: p, name: name, total: total}
}

type progressWriter struct {
	out         io.Writer
	display     *progressDisplay
	name        string
	total, done int64
}

func (w *progressWriter) Write(data []byte) (int, error) {
	n, err := w.out.Write(data)
	w.done += int64(n)
	if time.Since(w.display.last) >= 100*time.Millisecond || w.done >= w.total {
		percent := min(int(w.done*100/w.total), 100)
		filled := percent / 5
		w.display.draw(fmt.Sprintf("%s %3d%% [%s%s] %.1f/%.1f MiB", w.name, percent,
			strings.Repeat("=", filled), strings.Repeat(" ", 20-filled), float64(w.done)/(1<<20), float64(w.total)/(1<<20)))
		w.display.last = time.Now()
	}
	return n, err
}
