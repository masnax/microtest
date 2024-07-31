package main

import (
	"fmt"
	"sort"
	"time"

	cli "github.com/canonical/lxd/shared/cmd"
	"github.com/spf13/cobra"

	"github.com/masnax/microtest/v2/microcluster"
)

type cmdSecrets struct {
	common *CmdControl
}

func (c *cmdSecrets) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Manage join tokens for MicroCluster",
		RunE:  c.run,
	}

	var cmdAdd = cmdTokensAdd{common: c.common}
	cmd.AddCommand(cmdAdd.command())

	var cmdList = cmdTokensList{common: c.common}
	cmd.AddCommand(cmdList.command())

	var cmdRevoke = cmdTokensRevoke{common: c.common}
	cmd.AddCommand(cmdRevoke.command())

	return cmd
}

func (c *cmdSecrets) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

type cmdTokensAdd struct {
	common *CmdControl

	flagExpireAfter string
}

func (c *cmdTokensAdd) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new join token under the given name",
		RunE:  c.run,
	}
	cmd.Flags().StringVarP(&c.flagExpireAfter, "expire-after", "e", "3h", "Set the lifetime for the token")

	return cmd
}

func (c *cmdTokensAdd) run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir})
	if err != nil {
		return err
	}

	expireAfter, err := time.ParseDuration(c.flagExpireAfter)
	if err != nil {
		return fmt.Errorf("Invalid value for expire-after: %w", err)
	}

	token, err := m.NewJoinToken(cmd.Context(), args[0], expireAfter)
	if err != nil {
		return err
	}

	fmt.Println(token)

	return nil
}

type cmdTokensList struct {
	common *CmdControl

	flagFormat string
}

func (c *cmdTokensList) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List join tokens available for use",
		RunE:  c.run,
	}
	cmd.Flags().StringVarP(&c.flagFormat, "format", "f", cli.TableFormatTable, "Format (csv|json|table|yaml|compact)")

	return cmd
}

func (c *cmdTokensList) run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmd.Help()
	}

	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir})
	if err != nil {
		return err
	}

	records, err := m.ListJoinTokens(cmd.Context())
	if err != nil {
		return err
	}

	data := make([][]string, len(records))
	for i, record := range records {
		data[i] = []string{record.Name, record.Token, record.ExpiresAt.String()}
	}

	header := []string{"NAME", "TOKENS", "EXPIRES AT"}
	sort.Sort(cli.SortColumnsNaturally(data))

	return cli.RenderTable(c.flagFormat, header, data, records)
}

type cmdTokensRevoke struct {
	common *CmdControl
}

func (c *cmdTokensRevoke) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke <name>",
		Short: "Revoke the join token with the given name",
		RunE:  c.run,
	}

	return cmd
}

func (c *cmdTokensRevoke) run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir})
	if err != nil {
		return err
	}

	err = m.RevokeJoinToken(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return nil
}
