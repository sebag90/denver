package container

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// refreshInterval caps how often the live progress line is redrawn. The daemon
// emits an event per layer and per progress tick, which is far more often than
// a terminal can usefully show.
const refreshInterval = 80 * time.Millisecond

// pullProgress is one event of the /images/create stream. Docker and Podman
// both report downloaded bytes in progressDetail; the preformatted "progress"
// string is intentionally ignored so that the rendering stays ours.
type pullProgress struct {
	Status         string `json:"status"`
	ID             string `json:"id"`
	ProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`
	Error string `json:"error"`
}

func pullImage(client *http.Client, image string, w io.Writer) error {
	start := time.Now()

	req, err := http.NewRequest(http.MethodPost, baseApiUrl+"/images/create", nil)
	if err != nil {
		return fmt.Errorf("create pull request: %w", err)
	}

	q := req.URL.Query()
	q.Set("fromImage", image)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("pull image %s: %w", image, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pull image %s: %s: %s", image, resp.Status, bytes.TrimSpace(body))
	}

	reporter := newPullReporter(w)
	defer reporter.stop()

	state := newPullState()
	decoder := json.NewDecoder(resp.Body)
	for {
		var progress pullProgress
		if err := decoder.Decode(&progress); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("read pull progress: %w", err)
		}
		if progress.Error != "" {
			return fmt.Errorf("pull image %s: %s", image, progress.Error)
		}
		state.apply(progress)
		reporter.update(state.snapshot())
	}

	// Replace the live line with a durable one before reporting success.
	reporter.stop()
	fmt.Fprintf(w, "pulled %s (%s)\n", image, pullSummary(state.snapshot(), time.Since(start)))
	return nil
}

// pullLayer is the known progress of a single layer. Current only ever grows:
// each phase of a layer (download, extract) reports its own byte counter, and
// the aggregate should not jump backwards when a layer moves to the next phase.
type pullLayer struct {
	current int64
	total   int64
	done    bool
}

// pullState aggregates the per-layer events into a single view of the pull.
type pullState struct {
	layers map[string]*pullLayer
}

func newPullState() *pullState {
	return &pullState{layers: make(map[string]*pullLayer)}
}

func (s *pullState) apply(p pullProgress) {
	// Events without a layer ID carry tag or registry-level news ("Digest:
	// sha256:...", "Pulling from ..."), and statuses outside of layerStatus
	// are not part of the byte accounting.
	if p.ID == "" || !layerStatus(p.Status) {
		return
	}

	layer, ok := s.layers[p.ID]
	if !ok {
		layer = &pullLayer{}
		s.layers[p.ID] = layer
	}

	if p.ProgressDetail.Total > layer.total {
		layer.total = p.ProgressDetail.Total
	}
	if p.ProgressDetail.Current > layer.current {
		layer.current = p.ProgressDetail.Current
	}
	if layer.total > 0 && layer.current > layer.total {
		layer.current = layer.total
	}

	switch p.Status {
	case "Download complete", "Pull complete", "Already exists":
		if layer.total > 0 {
			layer.current = layer.total
		}
	}
	if p.Status == "Pull complete" || p.Status == "Already exists" {
		layer.done = true
	}
}

// pullSnapshot is the aggregate view handed to the reporter. Total is zero
// while the manifest is still being resolved and layer sizes are unknown.
type pullSnapshot struct {
	Current int64
	Total   int64
	Done    int
	Layers  int
}

func (s *pullState) snapshot() pullSnapshot {
	snap := pullSnapshot{Layers: len(s.layers)}
	for _, layer := range s.layers {
		snap.Current += layer.current
		snap.Total += layer.total
		if layer.done {
			snap.Done++
		}
	}
	return snap
}

func layerStatus(status string) bool {
	switch status {
	case "Pulling fs layer", "Waiting", "Downloading", "Verifying Checksum",
		"Download complete", "Extracting", "Pull complete", "Already exists":
		return true
	default:
		return false
	}
}

// pullSummary is the parenthesised tail of the final "pulled" line.
func pullSummary(s pullSnapshot, elapsed time.Duration) string {
	parts := make([]string, 0, 3)
	if s.Total > 0 {
		parts = append(parts, humanBytes(s.Total))
	}
	if s.Layers > 0 {
		unit := "layers"
		if s.Layers == 1 {
			unit = "layer"
		}
		parts = append(parts, fmt.Sprintf("%d %s", s.Layers, unit))
	}
	parts = append(parts, formatDuration(elapsed))
	return strings.Join(parts, ", ")
}

// pullReporter renders the stream of snapshots. It is either a live line
// redrawn in place (terminal) or a sequence of milestone lines (log file).
type pullReporter interface {
	update(pullSnapshot)
	stop()
}

func newPullReporter(w io.Writer) pullReporter {
	// TERM=dumb means the terminal cannot be relied on for cursor control.
	if f, ok := w.(*os.File); ok && term.IsTerminal(int(f.Fd())) && os.Getenv("TERM") != "dumb" {
		return &ttyReporter{w: f}
	}
	return &plainReporter{w: w}
}

// ttyReporter redraws a single aggregate line in place.
type ttyReporter struct {
	w        *os.File
	last     time.Time
	lastPct  int
	lastDone int
	frame    int
	drawn    bool
	stopped  bool
}

func (r *ttyReporter) update(s pullSnapshot) {
	now := time.Now()
	if r.drawn && s.Done == r.lastDone && now.Sub(r.last) < refreshInterval {
		return
	}
	r.last = now
	r.lastDone = s.Done
	r.drawn = true
	r.frame++

	line := progressLine(s, terminalWidth(r.w), &r.lastPct, r.frame)
	fmt.Fprintf(r.w, "\r\x1b[K%s", line)
}

func (r *ttyReporter) stop() {
	if r.stopped {
		return
	}
	r.stopped = true
	if r.drawn {
		fmt.Fprint(r.w, "\r\x1b[K")
	}
}

// plainReporter writes a line each time the pull crosses a decile, which keeps
// CI logs readable without the sub-second chatter of the raw event stream.
type plainReporter struct {
	w        io.Writer
	lastStep int
	stopped  bool
}

func (r *plainReporter) update(s pullSnapshot) {
	if s.Total <= 0 {
		return
	}
	pct := clampPercent(s.Current, s.Total)
	step := pct / 10
	if step <= r.lastStep {
		return
	}
	r.lastStep = step
	fmt.Fprintf(r.w, "  %3d%%  %s / %s%s\n", pct, humanBytes(s.Current), humanBytes(s.Total), layerSuffix(s))
}

func (r *plainReporter) stop() {
	r.stopped = true
}

// progressLine builds the live line, e.g.
//
//	[==============>--------]  62%  36.1 kB / 58.3 kB  7/12 layers
func progressLine(s pullSnapshot, width int, lastPct *int, frame int) string {
	spinner := "|/-\\"[frame%4]

	if s.Total <= 0 {
		return truncate(fmt.Sprintf("[ %c ] resolving%s", spinner, layerSuffix(s)), width)
	}

	pct := clampPercent(s.Current, s.Total)
	if pct < *lastPct {
		// Totals grow as layers are discovered, so the same byte count can
		// map to a smaller percentage later on. Keep moving forwards.
		pct = *lastPct
	}
	*lastPct = pct

	suffix := fmt.Sprintf(" %3d%%  %s / %s%s ", pct, humanBytes(s.Current), humanBytes(s.Total), layerSuffix(s))
	barWidth := width - len(suffix) - 2
	if barWidth < 10 {
		line := fmt.Sprintf("%3d%%  %s / %s%s", pct, humanBytes(s.Current), humanBytes(s.Total), layerSuffix(s))
		return truncate(line, width)
	}

	filled := barWidth * pct / 100
	bar := strings.Repeat("=", filled)
	if filled < barWidth {
		bar += ">" + strings.Repeat(" ", barWidth-filled-1)
	}
	return truncate("["+bar+"]"+suffix, width)
}

// truncate keeps the line narrower than the terminal, so that a wrapped line
// cannot break the in-place redraw.
func truncate(line string, width int) string {
	if width <= 1 || len(line) < width {
		return line
	}
	return line[:width-1]
}

func layerSuffix(s pullSnapshot) string {
	switch s.Layers {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("  %d/1 layer", s.Done)
	default:
		return fmt.Sprintf("  %d/%d layers", s.Done, s.Layers)
	}
}

func clampPercent(current, total int64) int {
	if total <= 0 {
		return 0
	}
	pct := int(current * 100 / total)
	if pct > 100 {
		pct = 100
	}
	return pct
}

func terminalWidth(w *os.File) int {
	width, _, err := term.GetSize(int(w.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func humanBytes(n int64) string {
	const unit = 1000
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n/div >= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "kMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
