package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Manage Hetzner Cloud SSH keys",
}

var sshListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all SSH keys in the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := client.ListSSHKeys()
		if err != nil {
			return err
		}

		if len(keys) == 0 {
			fmt.Println("No SSH keys found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tFINGERPRINT")
		fmt.Fprintln(w, "──\t────\t───────────")
		for _, k := range keys {
			fmt.Fprintf(w, "%d\t%s\t%s\n", k.ID, k.Name, k.Fingerprint)
		}
		return w.Flush()
	},
}

var sshAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Upload a new SSH key",
	Example: `  fornost ssh add --name my-key --key ~/.ssh/id_ed25519.pub`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		keyPath, _ := cmd.Flags().GetString("key")

		if name == "" || keyPath == "" {
			return fmt.Errorf("--name and --key are required")
		}

		data, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("read key file: %w", err)
		}

		publicKey := strings.TrimSpace(string(data))
		key, err := client.AddSSHKey(name, publicKey)
		if err != nil {
			return err
		}

		fmt.Printf("Created SSH key %q (ID: %d, Fingerprint: %s)\n", key.Name, key.ID, key.Fingerprint)
		return nil
	},
}

var sshDeleteCmd = &cobra.Command{
	Use:   "delete <name-or-id>",
	Short: "Delete an SSH key by name or ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteSSHKey(args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted SSH key %q\n", args[0])
		return nil
	},
}

func init() {
	sshAddCmd.Flags().String("name", "", "Name for the key in Hetzner")
	sshAddCmd.Flags().String("key", "", "Path to public key file")

	sshCmd.AddCommand(sshListCmd)
	sshCmd.AddCommand(sshAddCmd)
	sshCmd.AddCommand(sshDeleteCmd)
	rootCmd.AddCommand(sshCmd)
}
