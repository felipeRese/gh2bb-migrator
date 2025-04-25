package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	destURL string
	dryRun  bool
)

func main() {
	root := &cobra.Command{
		Use:   "gh2bb",
		Short: "Migrate a GitHub repo → Bitbucket by mirror-push over SSH",
		RunE:  run,
	}

	root.Flags().StringVar(&destURL, "dest-url", "",
		"SSH URL of Bitbucket repo to migrate to (e.g. git@bitbucket.org:workspace/name-of-the-repo.git)")
	root.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing")

	cobra.CheckErr(root.MarkFlagRequired("dest-url"))

	if err := root.Execute(); err != nil {
		logrus.Fatalf("Error: %v", err)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Validate destURL is SSH
	if !strings.HasPrefix(destURL, "git@") && !strings.HasPrefix(destURL, "ssh://") {
		return fmt.Errorf("dest-url must be SSH (e.g. git@bitbucket.org:workspace/name-of-the-repo.git)")
	}

	// Extract repo name from destURL
	// e.g. git@bitbucket.org:workspace/name-of-the-repo.git → name-of-the-repo
	re := regexp.MustCompile(`[:/](?P<workspace>[^/]+)/(?P<repo>[^/]+?)(?:\.git)?$`)
	matches := re.FindStringSubmatch(destURL)
	if matches == nil {
		return fmt.Errorf("could not parse repository name from dest-url")
	}
	repoName := matches[re.SubexpIndex("repo")]

	// Build the GitHub source URL
	sourceURL := fmt.Sprintf("git@github.com:be-growth/%s.git", repoName)

	logrus.Infof("Derived source-url: %s", sourceURL)

	// 1. Create a temp dir for the mirror clone
	tmpDir, err := os.MkdirTemp("", "gh2bb-")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneDir := filepath.Join(tmpDir, "repo.git")

	// 2. git clone --mirror (bare repo)
	if err := runCmd("git", dryRun, "", "clone", "--mirror", sourceURL, cloneDir); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// 3. In the bare repo, set the push URL of 'origin' to the Bitbucket destination
	if err := runCmd("git", dryRun, cloneDir,
		"remote", "set-url", "--push", "origin", destURL,
	); err != nil {
		return fmt.Errorf("git remote set-url failed: %w", err)
	}

	// 4. Push everything (branches, tags, refs) to Bitbucket via origin
	if err := runCmd("git", dryRun, cloneDir, "push", "--mirror", "origin"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	logrus.Infof("✅ Migration completed: %s → %s", sourceURL, destURL)
	return nil
}

// runCmd executes a command (e.g. git) with args in dir (if dir≠"").
// If dry is true, it just prints the command.
func runCmd(bin string, dry bool, dir string, args ...string) error {
	cmdStr := fmt.Sprintf("%s %s", bin, args)
	logrus.Infof("→ %s", cmdStr)

	if dry {
		return nil
	}

	cmd := exec.Command(bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = io.MultiWriter(os.Stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr)

	return cmd.Run()
}
