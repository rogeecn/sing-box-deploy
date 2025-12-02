package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/rogeecn/sing-box-deploy/internal/state"
	"github.com/spf13/cobra"
)

var (
	urlTagFilter  string
	urlTypeFilter string
)

var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "Print subscription URLs and optional QR codes",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := state.Load(getStatePath())
		if err != nil {
			if errors.Is(err, state.ErrNotFound) {
				return fmt.Errorf("state file not found, run deploy first")
			}
			return err
		}
		var matches []state.Inbound
		tagFilter := strings.ToLower(urlTagFilter)
		typeFilter := strings.ToLower(urlTypeFilter)
		for _, inbound := range st.Inbounds {
			if tagFilter != "" && !strings.Contains(strings.ToLower(inbound.Tag), tagFilter) {
				continue
			}
			if typeFilter != "" && strings.ToLower(inbound.Key) != typeFilter {
				continue
			}
			matches = append(matches, inbound)
		}
		if len(matches) == 0 {
			cmd.Println("no matching inbounds")
			return nil
		}
		for _, inbound := range matches {
			cmd.Printf("%s -> %s\n", inbound.Tag, inbound.ShareURL)
			escaped := url.QueryEscape(inbound.ShareURL)
			qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=%s", escaped)
			cmd.Printf("QR: %s\n\n", qrURL)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(urlCmd)
	urlCmd.Flags().StringVar(&urlTagFilter, "tag", "", "filter by inbound tag substring")
	urlCmd.Flags().StringVar(&urlTypeFilter, "type", "", "filter by inbound key (e.g. vless-ws-tls)")
}
