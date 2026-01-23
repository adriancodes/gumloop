package runner

import (
	"fmt"
	"time"
)

// Metrics tracks statistics about a run
type Metrics struct {
	Iterations int
	Commits    int
	StartTime  time.Time
	ExitReason string
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// Duration returns the elapsed time since the run started
func (m *Metrics) Duration() time.Duration {
	return time.Since(m.StartTime)
}

// FormatDuration formats a duration in "Xh Xm Xs" format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

// ExitReasonString returns a human-readable exit reason string
func ExitReasonString(code ExitCode) string {
	switch code {
	case ExitSuccess:
		return "✅ Complete (no changes)"
	case ExitError:
		return "❌ Error"
	case ExitSafety:
		return "⛔ Safety refusal"
	case ExitMaxIterations:
		return "🔄 Max iterations reached"
	case ExitStuck:
		return "🔁 Stuck (no commits)"
	case ExitInterrupt:
		return "⚠️  Interrupted"
	default:
		return fmt.Sprintf("Unknown exit code: %d", code)
	}
}
