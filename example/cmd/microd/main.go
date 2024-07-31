// Package microd provides the daemon.
package main

import (
	"context"
	"os"

	"github.com/canonical/lxd/shared/logger"
	"github.com/spf13/cobra"

	"github.com/masnax/microtest/v2/example/api"
	"github.com/masnax/microtest/v2/example/database"
	"github.com/masnax/microtest/v2/example/version"
	"github.com/masnax/microtest/v2/microcluster"
	"github.com/masnax/microtest/v2/rest/types"
	"github.com/masnax/microtest/v2/state"
)

// Debug indicates whether to log debug messages or not.
var Debug bool

// Verbose indicates verbosity.
var Verbose bool

type cmdGlobal struct {
	cmd *cobra.Command //nolint:structcheck,unused // FIXME: Remove the nolint flag when this is in use.

	flagHelp    bool
	flagVersion bool

	flagLogDebug   bool
	flagLogVerbose bool
}

func (c *cmdGlobal) run(cmd *cobra.Command, args []string) error {
	Debug = c.flagLogDebug
	Verbose = c.flagLogVerbose

	return logger.InitLogger("", "", c.flagLogVerbose, c.flagLogDebug, nil)
}

type cmdDaemon struct {
	global *cmdGlobal

	flagStateDir    string
	flagSocketGroup string
}

func (c *cmdDaemon) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microd",
		Short:   "Example daemon for MicroCluster - This will start a daemon with a running control socket and no database",
		Version: version.Version(),
	}

	cmd.RunE = c.run
	cmd.PreRunE = c.global.run

	return cmd
}

func (c *cmdDaemon) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{
		StateDir: c.flagStateDir,
	})
	if err != nil {
		return err
	}

	dargs := microcluster.DaemonArgs{
		Verbose: c.global.flagLogVerbose,
		Debug:   c.global.flagLogDebug,
		Version: version.Version(),

		SocketGroup: c.flagSocketGroup,

		ExtensionsSchema: database.SchemaExtensions,
		APIExtensions:    api.Extensions(),
		ExtensionServers: api.Servers,
	}

	// exampleHooks are some example post-action hooks that can be run by MicroCluster.
	dargs.Hooks = &state.Hooks{
		// PostBootstrap is run after the daemon is initialized and bootstrapped.
		PostBootstrap: func(ctx context.Context, s state.State, initConfig map[string]string) error {
			logCtx := logger.Ctx{}
			for k, v := range initConfig {
				logCtx[k] = v
			}

			// You can check your app extensions using the state.State object.
			hasMissingExt := s.HasExtension("missing_extension")
			if !hasMissingExt {
				logger.Warn("The 'missing_extension' is not registered")
			}

			// You can also check the internal extensions. (starting with "internal:" prefix)
			// These are read-only and defined at the MicroCluster level and cannot be added at runtime
			hasInternalExt := s.HasExtension("internal:runtime_extension_v1")
			if !hasInternalExt {
				logger.Warn("Every system should have the 'internal:runtime_extension_v1' extension")
			}

			logger.Info("This is a hook that runs after the daemon is initialized and bootstrapped")
			logger.Info("Here are the extra configuration keys that were passed into the init --bootstrap command", logCtx)

			return nil
		},

		PreBootstrap: func(ctx context.Context, s state.State, initConfig map[string]string) error {
			logCtx := logger.Ctx{}
			for k, v := range initConfig {
				logCtx[k] = v
			}

			logger.Info("This is a hook that runs before the daemon is initialized and bootstrapped")
			logger.Info("Here are the extra configuration keys that were passed into the init --bootstrap command", logCtx)

			return nil
		},

		// OnStart is run after the daemon is started.
		OnStart: func(ctx context.Context, s state.State) error {
			logger.Info("This is a hook that runs after the daemon first starts")

			return nil
		},

		// PostJoin is run after the daemon is initialized and joins a cluster.
		PostJoin: func(ctx context.Context, s state.State, initConfig map[string]string) error {
			logCtx := logger.Ctx{}
			for k, v := range initConfig {
				logCtx[k] = v
			}

			logger.Info("This is a hook that runs after the daemon is initialized and joins an existing cluster, after OnNewMember runs on all peers")
			logger.Info("Here are the extra configuration keys that were passed into the init --join command", logCtx)

			return nil
		},

		// PreJoin is run after the daemon is initialized and joins a cluster.
		PreJoin: func(ctx context.Context, s state.State, initConfig map[string]string) error {
			logCtx := logger.Ctx{}
			for k, v := range initConfig {
				logCtx[k] = v
			}

			logger.Info("This is a hook that runs after the daemon is initialized and joins an existing cluster, before OnNewMember runs on all peers")
			logger.Info("Here are the extra configuration keys that were passed into the init --join command", logCtx)

			return nil
		},

		// PostRemove is run after the daemon is removed from a cluster.
		PostRemove: func(ctx context.Context, s state.State, force bool) error {
			logger.Infof("This is a hook that is run on peer %q after a cluster member is removed, with the force flag set to %v", s.Name(), force)

			return nil
		},

		// PreRemove is run before the daemon is removed from the cluster.
		PreRemove: func(ctx context.Context, s state.State, force bool) error {
			logger.Infof("This is a hook that is run on peer %q just before it is removed, with the force flag set to %v", s.Name(), force)

			return nil
		},

		// OnHeartbeat is run after a successful heartbeat round.
		OnHeartbeat: func(ctx context.Context, s state.State) error {
			logger.Info("This is a hook that is run on the dqlite leader after a successful heartbeat")

			return nil
		},

		// OnNewMember is run after a new member has joined.
		OnNewMember: func(ctx context.Context, s state.State, newMember types.ClusterMemberLocal) error {
			logger.Infof("This is a hook that is run on peer %q when the new cluster member %q has joined", s.Name(), newMember.Name)

			return nil
		},

		// OnDaemonConfigUpdate is run after the local daemon config of a cluster member got modified.
		OnDaemonConfigUpdate: func(ctx context.Context, s state.State, config types.DaemonConfig) error {
			logger.Infof("Running OnDaemonConfigUpdate triggered by %q", config.Name)

			return nil
		},
	}

	return m.Start(cmd.Context(), dargs)
}

func main() {
	daemonCmd := cmdDaemon{global: &cmdGlobal{}}
	app := daemonCmd.command()
	app.SilenceUsage = true
	app.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	app.PersistentFlags().BoolVarP(&daemonCmd.global.flagHelp, "help", "h", false, "Print help")
	app.PersistentFlags().BoolVar(&daemonCmd.global.flagVersion, "version", false, "Print version number")
	app.PersistentFlags().BoolVarP(&daemonCmd.global.flagLogDebug, "debug", "d", false, "Show all debug messages")
	app.PersistentFlags().BoolVarP(&daemonCmd.global.flagLogVerbose, "verbose", "v", false, "Show all information messages")

	app.PersistentFlags().StringVar(&daemonCmd.flagStateDir, "state-dir", "", "Path to store state information"+"``")
	app.PersistentFlags().StringVar(&daemonCmd.flagSocketGroup, "socket-group", "", "Group to set socket's group ownership to")

	app.SetVersionTemplate("{{.Version}}\n")

	err := app.Execute()
	if err != nil {
		os.Exit(1)
	}
}
