package cmd

import (
	"errors"
	"fmt"
	"sort"

	"github.com/rogeecn/sing-box-deploy/internal/state"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show deployed inbound entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := state.Load(getStatePath())
		if err != nil {
			if errors.Is(err, state.ErrNotFound) {
				return fmt.Errorf("state file not found, run deploy first")
			}
			return err
		}
		cmd.Printf("Domain: %s\n", st.Domain)
		cmd.Printf("Subscription file: %s\n", st.SubscriptionFile)
		cmd.Println("Inbounds:")
		sort.Slice(st.Inbounds, func(i, j int) bool {
			return st.Inbounds[i].Tag < st.Inbounds[j].Tag
		})
		for _, inbound := range st.Inbounds {
			cmd.Printf("- %s [%s/%s] port:%d path:%s\n", inbound.Tag, inbound.Protocol, inbound.Transport, inbound.ListenPort, inbound.Path)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
