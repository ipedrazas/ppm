package cmd

import (
	"fmt"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

// addCollection writes a new collection entry (decision/question/task/note/
// conversation) from the resolved content.
func addCollection(cmd *cobra.Command, project string, t memory.EntryType, name string, extra []memory.KV) error {
	st, err := openStore()
	if err != nil {
		return err
	}
	content, err := contentInput.resolve()
	if err != nil {
		return err
	}
	entry, err := st.Write(project, t, content, memory.WriteOpts{Name: name, Extra: extra})
	if err != nil {
		return err
	}
	return emitWrote(cmd, entry)
}

// setSingleton replaces a project singleton (summary/focus) with new content.
func setSingleton(cmd *cobra.Command, project string, t memory.EntryType) error {
	st, err := openStore()
	if err != nil {
		return err
	}
	content, err := contentInput.resolve()
	if err != nil {
		return err
	}
	entry, err := st.Write(project, t, content, memory.WriteOpts{})
	if err != nil {
		return err
	}
	return emitWrote(cmd, entry)
}

func emitWrote(cmd *cobra.Command, entry *memory.Entry) error {
	return emit(cmd, output.Result{
		OK:      true,
		Message: fmt.Sprintf("Wrote %s → %s", entry.Type, entry.RelPath),
		Data:    entry,
	})
}
