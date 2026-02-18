package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Manage Hetzner Cloud firewalls",
}

var firewallListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all firewalls in the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		firewalls, err := client.ListFirewalls()
		if err != nil {
			return err
		}

		if len(firewalls) == 0 {
			fmt.Println("No firewalls found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tRULES\tAPPLIED TO")
		fmt.Fprintln(w, "──\t────\t─────\t──────────")
		for _, f := range firewalls {
			fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", f.ID, f.Name, len(f.Rules), len(f.AppliedTo))
		}
		return w.Flush()
	},
}

var firewallGetCmd = &cobra.Command{
	Use:   "get <name-or-id>",
	Short: "Get rules for a specific firewall",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fw, err := client.GetFirewall(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("ID:         %d\n", fw.ID)
		fmt.Printf("Name:       %s\n", fw.Name)
		fmt.Printf("Applied To: %d resources\n", len(fw.AppliedTo))
		fmt.Println()

		if len(fw.Rules) == 0 {
			fmt.Println("No rules configured.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "DIRECTION\tPROTOCOL\tPORT\tSOURCE IPs\tDESCRIPTION")
		fmt.Fprintln(w, "─────────\t────────\t────\t──────────\t───────────")
		for _, r := range fw.Rules {
			sourceIPs := strings.Join(r.SourceIPs, ", ")
			if r.Direction == "out" {
				sourceIPs = strings.Join(r.DestIPs, ", ")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				r.Direction,
				r.Protocol,
				r.Port,
				sourceIPs,
				r.Description,
			)
		}
		return w.Flush()
	},
}

func init() {
	firewallCmd.AddCommand(firewallListCmd)
	firewallCmd.AddCommand(firewallGetCmd)
	rootCmd.AddCommand(firewallCmd)
}
