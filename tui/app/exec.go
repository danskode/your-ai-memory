package app

import "os/exec"

// buildCmd creates an *exec.Cmd for use with tea.ExecProcess.
func buildCmd(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// buildCmdInDir creates an *exec.Cmd that runs in a specific working directory.
func buildCmdInDir(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd
}
