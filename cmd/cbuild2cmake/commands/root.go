/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"

	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Version string

var CopyrightNotice string

func printVersion(file io.Writer) {
	fmt.Fprintf(file, "cbuild2cmake version %v%v\n", Version, CopyrightNotice)
}

// UsageTemplate returns usage template for the command.
var usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Options:
{{.LocalFlags.FlagUsages | replaceString | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Options:
{{.InheritedFlags.FlagUsages | replaceString | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func preConfiguration(cmd *cobra.Command, args []string) error {
	// configure log level
	log.SetLevel(log.InfoLevel)
	debug, _ := cmd.Flags().GetBool("debug")
	quiet, _ := cmd.Flags().GetBool("quiet")
	logFile, _ := cmd.Flags().GetString("log")

	if debug {
		log.SetLevel(log.DebugLevel)
	} else if quiet {
		log.SetLevel(log.ErrorLevel)
	}
	if logFile != "" {
		parentLogDir := filepath.Dir(logFile)
		if _, err := os.Stat(parentLogDir); os.IsNotExist(err) {
			if err := os.MkdirAll(parentLogDir, 0755); err != nil {
				return err
			}
		}
		file, err := os.Create(logFile)
		if err != nil {
			return err
		}
		multiWriter := io.MultiWriter(os.Stdout, file)
		log.SetOutput(multiWriter)
	}
	return nil
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "cbuild2cmake [command] <name>.csolution.yml [options]",
		Short:             "cbuild2cmake: Generate CMakeLists " + Version + CopyrightNotice,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: preConfiguration,
		Args:              cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			versionFlag, _ := cmd.Flags().GetBool("version")
			if versionFlag {
				printVersion(cmd.OutOrStdout())
				return nil
			}

			var inputFile string
			if len(args) == 1 {
				inputFile = args[0]
			} else {
				_ = cmd.Help()
				return errors.New("invalid arguments")
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			debug, _ := cmd.Flags().GetBool("debug")
			verbose, _ := cmd.Flags().GetBool("verbose")
			clean, _ := cmd.Flags().GetBool("clean")

			options := maker.Options{
				Quiet:   quiet,
				Debug:   debug,
				Verbose: verbose,
				Clean:   clean,
			}

			configs, err := utils.GetInstallConfigs()
			if err != nil {
				return err
			}
			params := maker.Params{
				Runner:         utils.Runner{},
				Options:        options,
				InputFile:      inputFile,
				InstallConfigs: configs,
			}

			match, _ := regexp.MatchString(".*\\.cbuild-idx.yml", inputFile)
			if !match {
				return errors.New("invalid file argument")
			}

			log.Info("Generate CMakeLists " + Version + CopyrightNotice)
			m := &maker.Maker{Params: params}
			return m.GenerateCMakeLists()
		},
	}

	cobra.AddTemplateFunc("replaceString", func(s string) string {
		return strings.Replace(strings.Replace(s, "strings  ", "arg [...]", -1), "string ", "arg    ", -1)
	})
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.DisableFlagsInUseLine = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.Flags().BoolP("version", "V", false, "Print version")
	rootCmd.Flags().BoolP("help", "h", false, "Print usage")
	rootCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages except build invocations")
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug messages")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages from toolchain builds")
	rootCmd.Flags().BoolP("clean", "C", false, "Remove intermediate and output directories")

	rootCmd.SetFlagErrorFunc(FlagErrorFunc)
	//rootCmd.AddCommand(build.BuildCPRJCmd, list.ListCmd)
	return rootCmd
}

func FlagErrorFunc(cmd *cobra.Command, err error) error {
	if err != nil {
		log.Error(err)
		_ = cmd.Help()
	}
	return err
}
