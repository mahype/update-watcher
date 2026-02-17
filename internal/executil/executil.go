package executil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"time"
)

const defaultTimeout = 60 * time.Second

// Result holds the output of a command execution.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run executes a command with the default timeout and returns the result.
func Run(name string, args ...string) (*Result, error) {
	return RunWithTimeout(defaultTimeout, name, args...)
}

// RunWithTimeout executes a command with a custom timeout.
func RunWithTimeout(timeout time.Duration, name string, args ...string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %s", timeout)
	}

	return result, err
}

// RunAsSudo executes a command with sudo.
func RunAsSudo(name string, args ...string) (*Result, error) {
	sudoArgs := append([]string{name}, args...)
	return Run("sudo", sudoArgs...)
}

// RunAsUser executes a command as a specific user via sudo -u.
func RunAsUser(username string, name string, args ...string) (*Result, error) {
	current, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// If already running as the target user, run directly
	if current.Username == username {
		return Run(name, args...)
	}

	sudoArgs := append([]string{"-u", username, name}, args...)
	return Run("sudo", sudoArgs...)
}

// RunWithEnv executes a command with additional environment variables.
func RunWithEnv(env []string, name string, args ...string) (*Result, error) {
	return RunWithEnvTimeout(defaultTimeout, env, name, args...)
}

// RunWithEnvTimeout executes a command with additional environment variables and a custom timeout.
func RunWithEnvTimeout(timeout time.Duration, env []string, name string, args ...string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), env...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %s", timeout)
	}

	return result, err
}

// RunInDirWithEnv executes a command in a specific directory with additional environment variables.
func RunInDirWithEnv(dir string, env []string, name string, args ...string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %s", defaultTimeout)
	}

	return result, err
}

// RunAsUserWithEnv executes a command as a specific user with additional environment variables.
func RunAsUserWithEnv(env []string, username string, name string, args ...string) (*Result, error) {
	current, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// If already running as the target user, run directly with env
	if current.Username == username {
		return RunWithEnv(env, name, args...)
	}

	// Use "sudo -u <user> env VAR=val ... <command> <args>"
	sudoArgs := []string{"-u", username, "env"}
	sudoArgs = append(sudoArgs, env...)
	sudoArgs = append(sudoArgs, name)
	sudoArgs = append(sudoArgs, args...)
	return Run("sudo", sudoArgs...)
}
