package cmd

import (
	"bufio"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rogeecn/sing-box-deploy/internal/deployer"
	"github.com/rogeecn/sing-box-deploy/internal/spec"
	"github.com/spf13/cobra"
)

var (
	deployEmail   string
	deployTypes   []string
	deployRootDir string
	deployCaddy   string
	deploySubDir  string
	deployBinPath string
)

var deployCmd = &cobra.Command{
	Use:   "deploy <domain>",
	Short: "Render sing-box + Caddy configs for the given domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := strings.ToLower(strings.TrimSpace(args[0]))
		if domain == "" {
			return fmt.Errorf("domain is required")
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

		selectedTypes := deployTypes
		if len(selectedTypes) == 0 {
			choices, err := promptInboundSelection(cmd)
			if err != nil {
				return err
			}
			selectedTypes = choices
		}

		email := strings.TrimSpace(deployEmail)
		if email == "" {
			email = fmt.Sprintf("info@%s", domain)
		}
		opts := deployer.Options{
			Domain:          domain,
			Email:           email,
			InboundKeys:     selectedTypes,
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

		keySet := make(map[string]struct{}, len(selectedTypes))
		for _, k := range selectedTypes {
			keySet[k] = struct{}{}
		}
		cmd.Println("Share links:")
		for _, inbound := range st.Inbounds {
			if _, ok := keySet[inbound.Key]; !ok {
				continue
			}
			cmd.Printf("- %s -> %s\n", inbound.Tag, inbound.ShareURL)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVar(&deployEmail, "email", "", "email used for TLS certificate registration")
	deployCmd.Flags().StringSliceVar(&deployTypes, "type", nil, "inbound types to enable (repeatable)")
	deployCmd.Flags().StringVar(&deployRootDir, "root", "", "sing-box root directory (default /etc/sing-box)")
	deployCmd.Flags().StringVar(&deployCaddy, "caddy", "", "Caddyfile output path (default /etc/caddy/Caddyfile)")
	deployCmd.Flags().StringVar(&deploySubDir, "subscriptions", "", "directory for subscription files (default <root>/subscriptions)")
	deployCmd.Flags().StringVar(&deployBinPath, "sing-box-bin", "sing-box", "path to sing-box binary for helper commands")
}

func promptInboundSelection(cmd *cobra.Command) ([]string, error) {
	supported := spec.SupportedKeys()
	sort.Strings(supported)
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Available inbound templates:")
	for i, key := range supported {
		fmt.Fprintf(out, "  %d) %s\n", i+1, key)
	}
	fmt.Fprint(out, "请选择需要部署的协议编号(可用逗号分隔，默认全部): ")

	reader := bufio.NewReader(cmd.InOrStdin())
	line, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" || strings.EqualFold(line, "all") {
		return supported, nil
	}

	separators := []rune{',', '，', ' ', '\t', ';'}
	tokens := strings.FieldsFunc(line, func(r rune) bool {
		for _, sep := range separators {
			if r == sep {
				return true
			}
		}
		return false
	})
	if len(tokens) == 0 {
		return nil, fmt.Errorf("未选择任何协议")
	}
	indices := make(map[int]struct{})
	var selected []string
	for _, token := range tokens {
		idx, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("无效编号: %s", token)
		}
		if idx < 1 || idx > len(supported) {
			return nil, fmt.Errorf("编号超出范围: %d", idx)
		}
		if _, ok := indices[idx]; ok {
			continue
		}
		indices[idx] = struct{}{}
		selected = append(selected, supported[idx-1])
	}
	return selected, nil
}
