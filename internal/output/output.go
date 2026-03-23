package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Options struct {
	JSON bool
	JQ   string
}

func PrintJSON(w io.Writer, data interface{}, opts Options) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if opts.JQ != "" {
		return runJQ(w, payload, opts.JQ)
	}
	fmt.Fprintln(w, string(payload))
	return nil
}

func runJQ(w io.Writer, input []byte, expr string) error {
	if _, err := exec.LookPath("jq"); err != nil {
		fmt.Fprintln(os.Stderr, "jq not installed; printing JSON instead")
		fmt.Fprintln(w, string(input))
		return nil
	}
	cmd := exec.Command("jq", expr)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
