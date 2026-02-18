package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage Hetzner Cloud servers",
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all servers in the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		servers, err := client.ListServers()
		if err != nil {
			return err
		}

		if len(servers) == 0 {
			fmt.Println("No servers found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tIP\tTYPE\tLOCATION\tCREATED")
		fmt.Fprintln(w, "──\t────\t──────\t──\t────\t────────\t───────")
		for _, s := range servers {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
				s.ID,
				s.Name,
				s.Status,
				s.PublicNet.IPv4.IP,
				s.ServerType.Name,
				s.Datacenter.Location.Name,
				s.Created,
			)
		}
		return w.Flush()
	},
}

var serverGetCmd = &cobra.Command{
	Use:   "get <name-or-id>",
	Short: "Get details for a specific server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := client.GetServer(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("ID:         %d\n", server.ID)
		fmt.Printf("Name:       %s\n", server.Name)
		fmt.Printf("Status:     %s\n", server.Status)
		fmt.Printf("IPv4:       %s\n", server.PublicNet.IPv4.IP)
		fmt.Printf("Type:       %s\n", server.ServerType.Name)
		fmt.Printf("Datacenter: %s\n", server.Datacenter.Name)
		fmt.Printf("Location:   %s (%s)\n", server.Datacenter.Location.Name, server.Datacenter.Location.City)
		fmt.Printf("Created:    %s\n", server.Created)
		return nil
	},
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverGetCmd)
	rootCmd.AddCommand(serverCmd)
}
