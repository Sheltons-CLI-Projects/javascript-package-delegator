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
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// NewDevCmd creates a new Cobra command for the "dev" subcommand.
// It starts the development server after ensuring dependencies are installed.
func NewDevCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev [args...]",
		Short: "Run the dev script with automatic dependency checks",
		Long: `Dev runs the dev script after ensuring dependencies are installed
for the detected package manager. It checks for missing node_modules and
dependency changes before starting the development server.

Package Manager Behavior:
- npm:  Runs 'npm run dev [-- args]'
- pnpm: Runs 'pnpm run dev [-- args]'
- yarn: Runs 'yarn run dev [args]'
- bun:  Runs 'bun run dev [args]'
- deno: Runs 'deno task dev [args]'

Examples:
  jpd dev
  jpd dev -- --host 0.0.0.0
  jpd dev -- --port 3001`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			if pm == "" {
				return fmt.Errorf("no package manager detected for dev command")
			}

			noVoltaFlag, err := cmd.Flags().GetBool("no-volta")
			if err != nil {
				return fmt.Errorf("failed to parse --no-volta flag: %w", err)
			}

			targetDir, err := cmd.Flags().GetString(_CWD_FLAG)
			if err != nil {
				return fmt.Errorf("failed to parse --%s flag: %w", _CWD_FLAG, err)
			}
			if targetDir == "" {
				targetDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to determine working directory: %w", err)
				}
			}

			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			if err := autoInstallDependenciesIfNeeded(pm, "dev", targetDir, cmdRunner, goEnv, de, noVoltaFlag); err != nil {
				return err
			}

			var cmdArgs []string
			switch pm {
			case "npm", "pnpm":
				cmdArgs = []string{"run", "dev"}
				if len(args) > 0 {
					cmdArgs = append(cmdArgs, "--")
					cmdArgs = append(cmdArgs, args...)
				}
			case "yarn", "bun":
				cmdArgs = []string{"run", "dev"}
				cmdArgs = append(cmdArgs, args...)
			case "deno":
				if lo.Contains(args, "--eval") {
					return fmt.Errorf("don't pass --eval here use the exec command instead")
				}
				cmdArgs = []string{"task", "dev"}
				cmdArgs = append(cmdArgs, args...)
			default:
				return fmt.Errorf("dev command does not support package manager %q", pm)
			}

			cmdRunner.Command(pm, cmdArgs...)
			de.LogJSCommandIfDebugIsTrue(pm, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "pm", pm, "args", strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	cmd.Flags().Bool("no-volta", false, "Disable Volta integration during auto-install")

	return cmd
}
