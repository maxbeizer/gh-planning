package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Options struct {
	JSON bool
	JQ   string
}

func PrintJSON(data interface{}, opts Options) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if opts.JQ != "" {
		return runJQ(payload, opts.JQ)
	}
	fmt.Println(string(payload))
	return nil
}

func runJQ(input []byte, expr string) error {
	if _, err := exec.LookPath("jq"); err != nil {
		fmt.Fprintln(os.Stderr, "jq not installed; printing JSON instead")
		fmt.Println(string(input))
		return nil
	}
	cmd := exec.Command("jq", expr)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
