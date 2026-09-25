package main

import (
	"bytes"
	"strings"

	"github.com/brianmargolis/shades/client"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Cobra's fish script escapes the word being completed but evals the words
// before it as they are, so a finished 'theme;variant' argument splits at the
// ';' and the next Tab fails with "Unknown command" instead of completing.
const (
	fishArgsLine        = "set -l args (commandline -opc)"
	fishEscapedArgsLine = "set -l args (commandline -opc | string escape)"
)

// patchFishCompletion swaps the fish script out from under cobra's default
// completion command, leaving the other shells as cobra generates them.
func patchFishCompletion(root *cobra.Command) {
	root.InitDefaultCompletionCmd()
	fish, _, err := root.Find([]string{"completion", "fish"})
	if err != nil || fish.Name() != "fish" {
		return
	}

	fish.RunE = func(cmd *cobra.Command, args []string) error {
		noDescriptions, _ := cmd.Flags().GetBool("no-descriptions")
		script, err := fishScript(root, !noDescriptions)
		if err != nil {
			return err
		}
		_, err = cmd.OutOrStdout().Write([]byte(script))
		return err
	}
}

func fishScript(root *cobra.Command, includeDescriptions bool) (string, error) {
	buffer := bytes.Buffer{}
	if err := root.GenFishCompletion(&buffer, includeDescriptions); err != nil {
		return "", errors.Wrap(err, "generate fish completion")
	}
	return strings.Replace(buffer.String(), fishArgsLine, fishEscapedArgsLine, 1), nil
}

func completeThemes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	config, err := client.GetConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return config.Themes.Names(client.Filter{}), cobra.ShellCompDirectiveNoFileComp
}
