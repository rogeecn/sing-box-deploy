package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rogeecn/sing-box-deploy/internal/deployer"
	"github.com/spf13/cobra"
)

var (
	deployDomain  string
	deployEmail   string
	deployTypes   []string
	deployRootDir string
	deployCaddy   string
	deploySubDir  string
	deployBinPath string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Render sing-box + Caddy configs for the given domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(deployDomain) == "" {
			return fmt.Errorf("--domain is required")
		}
		rootDir := deployRootDir
		if rootDir == "" {
			rootDir = "/etc/sing-box"
		}
		subDir := deploySubDir
		if subDir == "" {
			subDir = filepath.Join(rootDir, "subscriptions")
		}
		caddyFile := deployCaddy
		if caddyFile == "" {
			caddyFile = "/etc/caddy/Caddyfile"
		}

		email := strings.TrimSpace(deployEmail)
		if email == "" {
			email = fmt.Sprintf("info@%s", strings.ToLower(deployDomain))
		}
		opts := deployer.Options{
			Domain:          strings.ToLower(deployDomain),
			Email:           email,
			InboundKeys:     deployTypes,
			RootDir:         rootDir,
			CaddyFile:       caddyFile,
			SubscriptionDir: subDir,
			StateFile:       getStatePath(),
			SingBoxBinary:   deployBinPath,
		}
		st, err := deployer.Run(opts)
		if err != nil {
			return err
		}
		cmd.Printf("Deployed %d inbounds for %s\n", len(st.Inbounds), st.Domain)
		cmd.Printf("sing-box config: %s\n", fmt.Sprintf("%s/00_common.json", st.RootDir))
		cmd.Printf("Caddyfile: %s\n", st.CaddyFile)
		cmd.Printf("Subscriptions: %s\n", st.SubscriptionFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(&deployDomain, "domain", "d", "", "domain to deploy (required)")
	deployCmd.Flags().StringVar(&deployEmail, "email", "", "email used for TLS certificate registration")
	deployCmd.Flags().StringSliceVar(&deployTypes, "type", nil, "inbound types to enable (repeatable)")
	deployCmd.Flags().StringVar(&deployRootDir, "root", "", "sing-box root directory (default /etc/sing-box)")
	deployCmd.Flags().StringVar(&deployCaddy, "caddy", "", "Caddyfile output path (default /etc/caddy/Caddyfile)")
	deployCmd.Flags().StringVar(&deploySubDir, "subscriptions", "", "directory for subscription files (default <root>/subscriptions)")
	deployCmd.Flags().StringVar(&deployBinPath, "sing-box-bin", "sing-box", "path to sing-box binary for helper commands")
}
