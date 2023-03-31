package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	fmt.Println("hello world")
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "build monorepo services",
		Long:  "build monorepo services",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔎 looking for Go services...")
			wd, err := os.Getwd()
			if err != nil {
				fmt.Println("💥 failed to get working directory:", err.Error())
				return
			}

			var directories []string
			filepath.Walk(wd, func(path string, info fs.FileInfo, _ error) error {
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

			if len(directories) == 0 {
				fmt.Println("no Go services found for build")
				return
			}

			fmt.Printf("👀 found %d services\n\n", len(directories))

			var wg sync.WaitGroup
			wg.Add(len(directories))

			for i := range directories {
				go func(dir string) {
					defer wg.Done()

					t0 := time.Now()

					s := strings.Split(dir, "/")
					serviceName := s[len(s)-1]

					envOut, err := exec.Command("go", "env").CombinedOutput()
					if err != nil {
						fmt.Println("💥 failed to get Go env:", err.Error())
						return
					}

					e := strings.ReplaceAll(string(envOut), "\"", "")
					goEnv := strings.Split(e, "\n")

					fmt.Printf("🏗  building service %s\n\n", serviceName)

					// we don't want to update the go.mod, but neither do we want to use it.

					cmd := exec.Command("sh", "-c", fmt.Sprintf("go build -mod readonly -o bin/%s", serviceName), dir)
					// cmd := exec.Command("go", "build", "-mod", "readonly", "-o", fmt.Sprintf("bin/%s", serviceName), dir)
					cmd.Dir = dir

					// Remove GOWORK so we can override
					for x := range goEnv {
						if strings.Contains(goEnv[x], "GOWORK") || strings.Contains(goEnv[x], "GO111MODULE") {
							goEnv[x] = ""
						}
					}

					// We need to set the environment for git config to fetch private repositories, etc.
					cmd.Env = append(cmd.Env, os.Environ()...)

					// Set the Go env.
					cmd.Env = append(cmd.Env, goEnv...)

					// Don't use workspaces.
					cmd.Env = append(cmd.Env, "GOWORK=off", "GO111MODULE=auto")

					b, err := cmd.CombinedOutput()
					if err != nil {
						fmt.Printf("⚠️  failed to build %s: %s\n", serviceName, string(b))
					} else {
						fmt.Printf("✅ %s built succesfully in %dms!\n\n", serviceName, time.Since(t0).Milliseconds())
					}
				}(directories[i])
			}

			wg.Wait()
		},
	}

	buildCmd.PersistentFlags().String("services", "", "Filter which services to build")

	rootCmd := &cobra.Command{Use: "hydra"}
	rootCmd.AddCommand(buildCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %s", err.Error())
	}
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
