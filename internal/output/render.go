// Package output renders command results in the agent-first JSON default or a
// human-readable text mode.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Format is the output encoding selected by the user.
type Format string

const (
	// JSON is the default, machine-first encoding.
	JSON Format = "json"
	// Text is the human-readable encoding.
	Text Format = "text"
)

// Result is the uniform envelope every command emits. Message is the
// human-facing rendering (mirrors the TS tools' text payload); Data carries the
// structured payload for the agent.
type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Render writes a successful result to w in the chosen format.
func Render(w io.Writer, f Format, r Result) error {
	if f == Text {
		if r.Message != "" {
			_, err := fmt.Fprintln(w, r.Message)
			return err
		}
		return nil
	}
	return encodeJSON(w, r)
}

// RenderError writes an error envelope. In JSON mode it goes to the same stdout
// channel as success (uniform parsing for the agent); callers signal failure via
// the process exit code.
func RenderError(w io.Writer, f Format, err error) {
	if f == Text {
		fmt.Fprintln(w, "Error:", err.Error())
		return
	}
	_ = encodeJSON(w, Result{OK: false, Error: err.Error()})
}

func encodeJSON(w io.Writer, r Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
