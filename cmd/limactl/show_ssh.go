package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/lima-vm/lima/pkg/sshutil"
	"github.com/lima-vm/lima/pkg/store"
	"github.com/spf13/cobra"
)

const showSSHExample = `
  "cmd" format (default): Full ssh command line.
    $ limactl show-ssh --format=cmd default
    ssh -o IdentityFile="/Users/example/.lima/_config/user" -o User=example -o Hostname=127.0.0.1 -o Port=60022 lima-default

  "args" format: Similar to the cmd format but omits "ssh" and the destination address
    $ limactl show-ssh --format=args default
    -o IdentityFile="/Users/example/.lima/_config/user" -o User=example -o Hostname=127.0.0.1 -o Port=60022

  "options" format: ssh option key value pairs
    $ limactl show-ssh --format=options default
    IdentityFile="/Users/example/.lima/_config/user"
    User=example
    Hostname=127.0.0.1
    Port=60022

  "config" format: ~/.ssh/config format
    $ limactl show-ssh --format=config default
    Host lima-default
      IdentityFile "/Users/example/.lima/_config/user "
      User example
      Hostname 127.0.0.1
      Port 60022

  To show the config file path:
    $ limactl ls --format='{{.SSHConfigFile}}' default
    /Users/example/.lima/default/ssh.config
`

func newShowSSHCommand() *cobra.Command {
	var shellCmd = &cobra.Command{
		Use:               "show-ssh [flags] INSTANCE",
		Short:             "Show the ssh command line",
		Example:           showSSHExample,
		Args:              WrapArgsError(cobra.ExactArgs(1)),
		RunE:              showSSHAction,
		ValidArgsFunction: showSSHBashComplete,
		SilenceErrors:     true,
	}

	shellCmd.Flags().StringP("format", "f", sshutil.FormatCmd, "Format: "+strings.Join(sshutil.Formats, ", "))
	_ = shellCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sshutil.Formats, cobra.ShellCompDirectiveNoFileComp
	})
	return shellCmd
}

func showSSHAction(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	instName := args[0]
	w := cmd.OutOrStdout()
	inst, err := store.Inspect(instName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("instance %q does not exist, run `limactl start %s` to create a new instance", instName, instName)
		}
		return err
	}
	y, err := inst.LoadYAML()
	if err != nil {
		return err
	}
	opts, err := sshutil.SSHOpts(inst.Dir, *y.SSH.LoadDotSSHPubKeys, *y.SSH.ForwardAgent, *y.SSH.ForwardX11, *y.SSH.ForwardX11Trusted)
	if err != nil {
		return err
	}
	opts = append(opts, "Hostname=127.0.0.1")
	opts = append(opts, fmt.Sprintf("Port=%d", inst.SSHLocalPort))
	return sshutil.Format(w, instName, format, opts)
}

func showSSHBashComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return bashCompleteInstanceNames(cmd)
}
