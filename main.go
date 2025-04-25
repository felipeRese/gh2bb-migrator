package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	destURL  string
	dryRun   bool
	ghPrefix string
)

func main() {
	godotenv.Load()
	ghPrefix = os.Getenv("GH_PREFIX")
	if ghPrefix == "" {
		logrus.Fatal("GH_PREFIX must be set in .env or environment")
	}

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
	if !strings.HasPrefix(destURL, "git@") && !strings.HasPrefix(destURL, "ssh://") {
		return fmt.Errorf("dest-url must be SSH (e.g. git@bitbucket.org:workspace/name-of-the-repo.git)")
	}

	re := regexp.MustCompile(`[:/][^/]+/(?P<repo>[^/]+?)(?:\.git)?$`)
	matches := re.FindStringSubmatch(destURL)
	if matches == nil {
		return fmt.Errorf("could not parse repository name from dest-url")
	}
	repoName := matches[re.SubexpIndex("repo")]

	sourceURL := fmt.Sprintf("git@github.com:%s/%s.git", ghPrefix, repoName)
	logrus.Infof("Derived source-url: %s", sourceURL)

	tmpDir, err := os.MkdirTemp("", "gh2bb-")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneDir := filepath.Join(tmpDir, "repo.git")

	if err := runCmd("git", dryRun, "", "clone", "--mirror", sourceURL, cloneDir); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	if err := runCmd("git", dryRun, cloneDir, "remote", "set-url", "--push", "origin", destURL); err != nil {
		return fmt.Errorf("git remote set-url failed: %w", err)
	}

	if err := runCmd("git", dryRun, cloneDir, "push", "--mirror", "origin"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	logrus.Infof("✅ Migration completed: %s → %s", sourceURL, destURL)
	return nil
}

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
