/*
Copyright © 2025 Shelton Louis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package cmd provides command-line interface implementations for the JavaScript package delegator.
package cmd

import (
	// standard library
	"fmt"
	"strings"

	// external
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates a new Cobra command for the "create" subcommand.
// It delegates directly to the package manager's native create command.
func NewCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [args...]",
		Short: "Scaffold a new project using the package manager's create command",
		Long: `Scaffold a new project by delegating to the package manager's native create command.

Package Manager Behavior:
- npm:  Runs 'npm create <args>'   (npm create = npm init, runs create-* packages)
- pnpm: Runs 'pnpm create <args>'
- yarn: Runs 'yarn create <args>'  (works for both yarn v1 and v2+)
- bun:  Runs 'bun create <args>'
- deno: Runs 'deno create <args>'

Examples:
  jpd create vite@latest my-app
  jpd create next-app myapp --typescript
  jpd -a deno create npm:vite my-deno-app`,
		Aliases: []string{"c"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Using package manager", "pm", pm)
			})

			var execCommand string
			var cmdArgs []string

			switch pm {
			case "npm", "pnpm", "yarn", "bun":
				execCommand = pm
				cmdArgs = append([]string{"create"}, args...)
			case "deno":
				execCommand = "deno"
				cmdArgs = append([]string{"create"}, args...)
			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			de.LogJSCommandIfDebugIsTrue(execCommand, cmdArgs...)
			cmdRunner.Command(execCommand, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "cmd", execCommand, "args", strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}
}
