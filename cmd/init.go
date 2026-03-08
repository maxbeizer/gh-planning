package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/spf13/cobra"
)

var initOpts struct {
	Project int
	Owner   string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gh-planning configuration",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().IntVar(&initOpts.Project, "project", 0, "Default project number")
	initCmd.Flags().StringVar(&initOpts.Owner, "owner", "", "Project owner (defaults to authenticated user)")
}

func runInit(cmd *cobra.Command, args []string) error {
	owner := initOpts.Owner
	if owner == "" {
		current, err := github.CurrentUser(cmd.Context())
		if err != nil {
			return err
		}
		owner = current
	}
	project := initOpts.Project
	if project == 0 {
		fmt.Print("Project number: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)
		value, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Errorf("invalid project number")
		}
		project = value
	}
	if _, err := github.VerifyProject(cmd.Context(), owner, project); err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.DefaultOwner = owner
	cfg.DefaultProject = project
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Saved config: owner=%s project=%d\n", owner, project)
	return nil
}
