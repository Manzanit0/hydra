package tool

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetEnv get Go environment, i.e. `go env`.
func GetEnv() ([]string, error) {
	envOut, err := exec.Command("go", "env").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exec go env: %w", err)
	}

	// FIXME: this probably doesn't work on windows?
	e := strings.ReplaceAll(string(envOut), "\"", "")
	goEnv := strings.Split(e, "\n")
	return goEnv, nil
}

// Build build the Go module in dir and output it with name out.
func Build(out, dir string) (string, error) {
	goEnv, err := GetEnv()
	if err != nil {
		return "", fmt.Errorf("get Go ENV: %w", err)
	}

	cmd := exec.Command("go", "build", "-mod", "readonly", "-o", filepath.Join("bin", out), dir)

	// Run it local to the module we want to compile, not in the parent directory.
	cmd.Dir = dir

	// Remove GOWORK so we can override later.
	for x := range goEnv {
		if strings.Contains(goEnv[x], "GOWORK") || strings.Contains(goEnv[x], "GO111MODULE") {
			goEnv[x] = ""
		}
	}

	// We need to set the environment for git config to fetch private repositories, etc.
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, goEnv...)
	// Don't use workspaces as it can have side-effects depending on user local config.
	cmd.Env = append(cmd.Env, "GOWORK=off", "GO111MODULE=auto")

	b, err := cmd.CombinedOutput()
	if err != nil {
		return string(b), fmt.Errorf("build module: %w", err)
	}

	return string(b), nil
}

func Tidy(dir string) (string, error) {
	goEnv, err := GetEnv()
	if err != nil {
		return "", fmt.Errorf("get Go ENV: %w", err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, goEnv...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return string(b), fmt.Errorf("tidy module: %w", err)
	}

	return string(b), nil
}

func Vendor(dir string) (string, error) {
	goEnv, err := GetEnv()
	if err != nil {
		return "", fmt.Errorf("get Go ENV: %w", err)
	}

	cmd := exec.Command("go", "mod", "vendor")
	cmd.Dir = dir

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, goEnv...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return string(b), fmt.Errorf("vendor module: %w", err)
	}

	return string(b), nil
}

func Test(dir string) (string, error) {
	goEnv, err := GetEnv()
	if err != nil {
		return "", fmt.Errorf("get Go ENV: %w", err)
	}

	cmd := exec.Command("go", "test", "-race", "-shuffle=on", "./...")

	// Run it local to the module we want to compile, not in the parent directory.
	cmd.Dir = dir

	// Remove GOWORK so we can override later.
	for x := range goEnv {
		if strings.Contains(goEnv[x], "GOWORK") {
			goEnv[x] = ""
		}
	}

	// We need to set the environment for git config to fetch private repositories, etc.
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, goEnv...)
	// Don't use workspaces as it can have side-effects depending on user local config.
	cmd.Env = append(cmd.Env, "GOWORK=off")

	b, err := cmd.CombinedOutput()
	if err != nil {
		return string(b), fmt.Errorf("test module: %w", err)
	}

	return string(b), nil
}

// FindGoServices recursively look for Go services under all directories in the
// provided working directory.
func FindGoServices(workingDir string) ([]string, error) {
	var directories []string
	err := filepath.Walk(workingDir, func(path string, info fs.FileInfo, _ error) error {
		if info.IsDir() && info.Name() == "vendor" {
			return filepath.SkipDir
		}

		if !info.IsDir() && info.Name() == "go.mod" {
			dir := filepath.Dir(path)
			if e, _ := exists(filepath.Join(dir, "main.go")); e {
				directories = append(directories, dir)
			}
		}

		return nil
	})

	return directories, err
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
