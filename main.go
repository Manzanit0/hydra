package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/manzanit0/hydra/pkg/tool"
	"github.com/spf13/cobra"
)

func main() {
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

			directories, err := tool.FindGoServices(wd)
			if err != nil {
				fmt.Println("💥 failed to find Go services:", err.Error())
				return
			}

			if len(directories) == 0 {
				fmt.Println("no Go services found for build")
				return
			}

			fmt.Printf("👀 found %d services\n", len(directories))

			var filters []string
			if services := cmd.Flag("services").Value.String(); services != "" {
				fmt.Println("building only:", services)
				filters = strings.Split(services, ",")
			}

			var wg sync.WaitGroup

			for i := range directories {
				// If there are filters, then only build the provided ones.
				// Otherwise, build all.
				if len(filters) > 0 {
					s := strings.Split(directories[i], "/")
					serviceName := s[len(s)-1]

					doBuild := false
					for j := range filters {
						if filters[j] == serviceName {
							doBuild = true
							break
						}
					}

					if !doBuild {
						continue
					}
				}

				wg.Add(1)
				go func(dir string) {
					defer wg.Done()

					t0 := time.Now()

					s := strings.Split(dir, "/")
					serviceName := s[len(s)-1]

					fmt.Printf("🏗  building service %s\n\n", serviceName)

					out, err := tool.Build(serviceName, dir)
					if err != nil {
						fmt.Printf("⚠️  failed to build %s: %s\n", serviceName, out)
					} else {
						fmt.Printf("✅ %s built succesfully in %dms!\n\n", serviceName, time.Since(t0).Milliseconds())
					}
				}(directories[i])
			}

			wg.Wait()
		},
	}

	buildCmd.PersistentFlags().String("services", "", "Filter which services to build")

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "test monorepo services",
		Long:  "test monorepo services",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔎 looking for Go services...")
			wd, err := os.Getwd()
			if err != nil {
				fmt.Println("💥 failed to get working directory:", err.Error())
				return
			}

			directories, err := tool.FindGoServices(wd)
			if err != nil {
				fmt.Println("💥 failed to find Go services:", err.Error())
				return
			}

			if len(directories) == 0 {
				fmt.Println("no Go services found for build")
				return
			}

			fmt.Printf("👀 found %d services\n\n", len(directories))

			var wg sync.WaitGroup
			wg.Add(len(directories))
			// TODO: I wonder if this is the best value possible for concurrent
			// test suites being run?
			sem := make(chan bool, runtime.GOMAXPROCS(0))

			for i := range directories {
				go func(dir string) {
					defer wg.Done()
					defer func() {
						<-sem
					}()

					sem <- true

					t0 := time.Now()

					s := strings.Split(dir, "/")
					serviceName := s[len(s)-1]

					// TODO: this message makes it confusing do to the async
					// nature of the messages showing.
					// fmt.Printf("🧪 launching tests for %s\n\n", serviceName)

					// TODO: I don't think it's a good idea to run tidy and
					// vendor always, but I need to understand more how to
					// identify _when_ to run them.
					_, _ = tool.Tidy(dir)
					_, _ = tool.Vendor(dir)

					out, err := tool.Test(dir)
					if err != nil {
						fmt.Printf("⚠️  failed to test %s:\n%s\n", serviceName, out)
					} else {
						fmt.Printf("✅ %s test succesfully in %dms!\n%s\n", serviceName, time.Since(t0).Milliseconds(), out)
					}
				}(directories[i])
			}

			wg.Wait()
		},
	}

	rootCmd := &cobra.Command{Use: "hydra"}
	rootCmd.AddCommand(buildCmd, testCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %s", err.Error())
	}
}
