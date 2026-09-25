package main

import (
	"fmt"

	"github.com/brianmargolis/shades/client"
	"github.com/brianmargolis/shades/picker"
	"github.com/brianmargolis/shades/preview"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "shades",
		Short: "Synchronize color themes across your terminal tools",
		Long: `shades synchronizes color themes across your terminal tools.

With no command, shades toggles between your default themes. Everything but
install and uninstall talks to the server, so 'shades server' has to be running
(which 'shades install' sets up).

Config: $SHADES_CONFIG, else ~/.config/shades/shades.yaml
State:  $SHADES_STATE, else ~/.shades/state.yaml
Logs:   ~/.shades/logs`,
		Example: `  shades set 'everforest;dark-medium'
  shades i --dark
  shades clients tmux ghostty bat`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initLogger()
			zap.S().Infow("starting shades", "command", cmd.CommandPath(), "args", args)
		},
		RunE: withConfig(runToggle),
	}

	root.AddCommand(
		&cobra.Command{
			Use:     "dark",
			Aliases: []string{"d"},
			Short:   "Switch to the default dark theme",
			Args:    cobra.NoArgs,
			RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
				return client.ChangerClient{Theme: config.DefaultDarkTheme}.Start(cmd.Context(), socketPath)
			}),
		},
		&cobra.Command{
			Use:     "light",
			Aliases: []string{"l"},
			Short:   "Switch to the default light theme",
			Args:    cobra.NoArgs,
			RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
				return client.ChangerClient{Theme: config.DefaultLightTheme}.Start(cmd.Context(), socketPath)
			}),
		},
		&cobra.Command{
			Use:     "toggle",
			Aliases: []string{"t"},
			Short:   "Switch to the default theme on the other side of the current one",
			Args:    cobra.NoArgs,
			RunE:    withConfig(runToggle),
		},
		&cobra.Command{
			Use:   "set <theme;variant>",
			Short: "Switch to a specific theme",
			Args:  themeArg,
			RunE: func(cmd *cobra.Command, args []string) error {
				return client.ChangerClient{Theme: args[0]}.Start(cmd.Context(), socketPath)
			},
		},
		newRandomCommand(),
		newInteractiveCommand(),
		&cobra.Command{
			Use:     "preview <theme;variant>",
			Aliases: []string{"p"},
			Short:   "Print a theme's palette as swatches",
			Args:    themeArg,
			RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
				variant, err := config.Themes.GetVariant(args[0])
				if err != nil {
					return err
				}
				swatches, err := preview.NewPreviewer().Preview(variant)
				if err != nil {
					return err
				}
				fmt.Println(swatches)
				return nil
			}),
		},
		newGalleryCommand(),
		&cobra.Command{
			Use:     "favorite",
			Aliases: []string{"fav"},
			Short:   "Mark the current theme as a favorite",
			Args:    cobra.NoArgs,
			RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
				return runFavorite(config, true)
			}),
		},
		&cobra.Command{
			Use:     "unfavorite",
			Aliases: []string{"unfav"},
			Short:   "Unmark the current theme as a favorite",
			Args:    cobra.NoArgs,
			RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
				return runFavorite(config, false)
			}),
		},
		&cobra.Command{
			Use:   "default",
			Short: "Make the current theme the dark or light default",
			Args:  cobra.NoArgs,
			RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
				return runDefault(config)
			}),
		},
		&cobra.Command{
			Use:   "state",
			Short: "Print what shades has saved over your config",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runState()
			},
		},
		newListCommand(),
		&cobra.Command{
			Use:   "server",
			Short: "Run the server, which every other command talks to",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return NewServer().Start(socketPath)
			},
		},
		newClientsCommand(),
		&cobra.Command{
			Use:   "install",
			Short: "Install the server and clients as launchd agents",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runInstall()
			},
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Remove the launchd agents",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runUninstall()
			},
		},
		newPickerListCommand(),
		&cobra.Command{
			Use:    picker.ToggleFavoriteCommand + " <theme;variant>",
			Hidden: true,
			Args:   themeArg,
			RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
				return runToggleFavorite(config, args[0])
			}),
		},
	)

	return root
}

func newRandomCommand() *cobra.Command {
	filter := client.Filter{}
	command := &cobra.Command{
		Use:   "random",
		Short: "Switch to a random theme",
		Args:  cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
			return runRandom(cmd.Context(), config, filter)
		}),
	}
	addFilterFlags(command, &filter)
	return command
}

func newInteractiveCommand() *cobra.Command {
	opts := picker.PickerOpts{SocketPath: socketPath}
	command := &cobra.Command{
		Use:     "interactive",
		Aliases: []string{"i"},
		Short:   "Pick a theme in an fzf window, with live preview",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := picker.NewPicker().Start(opts)
			return err
		},
	}
	addFilterFlags(command, &opts.Filter)
	command.Flags().BoolVar(&opts.UseTmux, "tmux", false, "use fzf-tmux, a floating tmux window")
	return command
}

func newGalleryCommand() *cobra.Command {
	filter := client.Filter{}
	command := &cobra.Command{
		Use:   "gallery",
		Short: "Print every palette at once, one row per variant",
		Args:  cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
			runGallery(config, filter)
			return nil
		}),
	}
	addFilterFlags(command, &filter)
	return command
}

func newListCommand() *cobra.Command {
	filter := client.Filter{}
	command := &cobra.Command{
		Use:   "list",
		Short: "List every theme;variant in your config",
		Args:  cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
			for _, name := range config.Themes.Names(filter) {
				fmt.Println(name)
			}
			return nil
		}),
	}
	addFilterFlags(command, &filter)
	return command
}

func newClientsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clients <client>...",
		Short: "Run one or more clients in the foreground",
		Long:  "Run one or more clients in the foreground.\n\nClients:\n" + clientList(),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("clients needs at least one client name; run 'shades clients --help' to see them")
			}
			clients := newClients()
			for _, name := range args {
				if _, ok := clients[name]; !ok {
					return errors.Errorf("no such client %q; run 'shades clients --help' to see them", name)
				}
			}
			return nil
		},
		ValidArgs: clientNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClients(args)
		},
	}
}

func newPickerListCommand() *cobra.Command {
	filter := client.Filter{}
	command := &cobra.Command{
		Use:    picker.ListCommand,
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, config client.ConfigModel, args []string) error {
			for _, line := range picker.Lines(config, filter) {
				fmt.Println(line)
			}
			return nil
		}),
	}
	addFilterFlags(command, &filter)
	return command
}

func addFilterFlags(command *cobra.Command, filter *client.Filter) {
	command.Flags().BoolVarP(&filter.OnlyDark, "dark", "d", false, "only dark variants")
	command.Flags().BoolVarP(&filter.OnlyLight, "light", "l", false, "only light variants")
	command.Flags().BoolVarP(&filter.OnlyFavorites, "favorites", "f", false, "only favorite variants")
	command.MarkFlagsMutuallyExclusive("dark", "light")
}

// withConfig loads the config for commands that need it, so the ones that
// don't (install, help, completion) still work without one.
func withConfig(
	run func(cmd *cobra.Command, config client.ConfigModel, args []string) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		config, err := client.GetConfig()
		if err != nil {
			return errors.Wrap(err, "could not load config")
		}
		zap.S().Debugw("config", "config", config)
		return run(cmd, config, args)
	}
}

// themeArg is spelled out rather than cobra.ExactArgs(1) because the likely
// mistake is an unquoted name, which the shell splits at the ';'.
func themeArg(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.Errorf(
			"%s needs one <theme;variant> argument, quoted so the shell doesn't split it at the ';'; run 'shades list' to see them",
			cmd.Name(),
		)
	}
	return nil
}

func runToggle(cmd *cobra.Command, config client.ConfigModel, args []string) error {
	toggler := client.TogglerClient{
		DarkTheme:  config.DefaultDarkTheme,
		LightTheme: config.DefaultLightTheme,
		Themes:     config.Themes,
	}
	return errors.Wrap(toggler.Start(cmd.Context(), socketPath), "could not toggle the theme")
}

// runClients runs each client until they all finish, and stops at the first
// one that fails.
func runClients(names []string) error {
	clients := newClients()
	results := make(chan error, len(names))
	for _, name := range names {
		go func() {
			results <- errors.Wrapf(clients[name].Start(socketPath), "client %s failed", name)
		}()
	}

	for range names {
		if err := <-results; err != nil {
			return err
		}
	}
	return nil
}
