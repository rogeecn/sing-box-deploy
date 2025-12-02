package deployer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rogeecn/sing-box-deploy/internal/share"
	"github.com/rogeecn/sing-box-deploy/internal/spec"
	"github.com/rogeecn/sing-box-deploy/internal/state"
	"github.com/rogeecn/sing-box-deploy/internal/templates"
)

type Options struct {
	Domain          string
	Email           string
	InboundKeys     []string
	RootDir         string
	CaddyFile       string
	SubscriptionDir string
	StateFile       string
	SingBoxBinary   string
	TLSKeyPath      string
	TLSCertPath     string
}

func (o *Options) validate() error {
	if o.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if o.RootDir == "" {
		return fmt.Errorf("root directory is required")
	}
	if o.CaddyFile == "" {
		return fmt.Errorf("caddy file is required")
	}
	if o.SubscriptionDir == "" {
		return fmt.Errorf("subscription directory is required")
	}
	if o.StateFile == "" {
		return fmt.Errorf("state file path is required")
	}
	if o.SingBoxBinary == "" {
		o.SingBoxBinary = "sing-box"
	}
	if o.TLSKeyPath == "" {
		o.TLSKeyPath = filepath.Join(o.RootDir, "tls.key")
	}
	if o.TLSCertPath == "" {
		o.TLSCertPath = filepath.Join(o.RootDir, "tls.cer")
	}
	return nil
}

// Run executes the deployment workflow and returns the resulting state.
func Run(opts Options) (*state.State, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	keys, err := spec.NormalizeKeys(opts.InboundKeys)
	if err != nil {
		return nil, err
	}

	inbounds := make(map[string]spec.InboundSpec, len(keys))
	for _, key := range keys {
		specData, err := spec.BuildSpec(key, opts.Domain)
		if err != nil {
			return nil, err
		}
		inbounds[key] = specData
	}

	data := templates.Data{
		Domain:      opts.Domain,
		Email:       opts.Email,
		Inbounds:    inbounds,
		TLSKeyPath:  opts.TLSKeyPath,
		TLSCertPath: opts.TLSCertPath,
	}

	rendered, err := templates.RenderInbounds(data)
	if err != nil {
		return nil, err
	}

	if err := writeInboundFiles(opts.RootDir, inbounds, rendered); err != nil {
		return nil, err
	}

	if err := writeSingBoxConfig(opts.RootDir); err != nil {
		return nil, err
	}

	if err := ensureTLSKeyPair(opts); err != nil {
		return nil, err
	}

	if err := writeCaddyFile(opts.CaddyFile, data); err != nil {
		return nil, err
	}

	shareLinks := make([]state.Inbound, 0, len(keys))
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# Subscriptions for %s\n\n", opts.Domain))

	for _, key := range keys {
		specData := inbounds[key]
		link, err := share.BuildLink(specData, opts.Domain)
		if err != nil {
			return nil, err
		}
		shareLinks = append(shareLinks, state.Inbound{
			Key:        key,
			Tag:        specData.Tag,
			Protocol:   specData.Protocol,
			Transport:  specData.Transport,
			ListenPort: specData.ListenPort,
			UUID:       specData.UUID,
			Path:       specData.Path,
			Host:       specData.Host,
			ShareURL:   link,
		})
		builder.WriteString(fmt.Sprintf("[%s]\n%s\n\n", specData.Tag, link))
	}

	subPath, err := writeSubscription(opts.SubscriptionDir, opts.Domain, builder.String())
	if err != nil {
		return nil, err
	}

	finalState := &state.State{
		Domain:           opts.Domain,
		Email:            opts.Email,
		RootDir:          opts.RootDir,
		CaddyFile:        opts.CaddyFile,
		SubscriptionFile: subPath,
		Inbounds:         shareLinks,
	}
	if err := state.Save(opts.StateFile, finalState); err != nil {
		return nil, err
	}
	return finalState, nil
}

func writeInboundFiles(root string, specs map[string]spec.InboundSpec, rendered map[string][]byte) error {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return err
	}
	for key, content := range rendered {
		specData := specs[key]
		file := filepath.Join(root, "02_inbounds_"+specData.FileName)
		var inbound map[string]any
		if err := json.Unmarshal(content, &inbound); err != nil {
			return fmt.Errorf("decode inbound %s: %w", key, err)
		}
		wrapper := map[string]any{
			"inbounds": []any{inbound},
		}
		encoded, err := json.MarshalIndent(wrapper, "", "  ")
		if err != nil {
			return fmt.Errorf("wrap inbound %s: %w", key, err)
		}
		if err := os.WriteFile(file, append(encoded, '\n'), 0o640); err != nil {
			return fmt.Errorf("write %s: %w", file, err)
		}
	}
	return nil
}

func writeSingBoxConfig(root string) error {
	configPath := filepath.Join(root, "00_common.json")
	payload := map[string]any{
		"log": map[string]any{
			"level":     "info",
			"timestamp": true,
			"output":    "/var/log/sing-box/sing-box.log",
		},
		"outbounds": []map[string]any{
			{
				"type": "direct",
				"tag":  "direct",
			},
			{
				"type": "block",
				"tag":  "block",
			},
		},
		"route": map[string]any{
			"final": "direct",
		},
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return err
	}
	return os.WriteFile(configPath, append(data, '\n'), 0o640)
}

func ensureTLSKeyPair(opts Options) error {
	keyPath := opts.TLSKeyPath
	certPath := opts.TLSCertPath
	certDir := filepath.Dir(keyPath)
	if _, err := os.Stat(keyPath); err == nil {
		if _, err := os.Stat(certPath); err == nil {
			return nil
		}
	}
	if err := os.MkdirAll(certDir, 0o750); err != nil {
		return fmt.Errorf("create cert dir: %w", err)
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(opts.SingBoxBinary, "generate", "tls-keypair", opts.Domain, "-m", "1024")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("generate tls keypair: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	keyPEM, err := extractPEM(stdout.String(), "PRIVATE KEY")
	if err != nil {
		return err
	}
	certPEM, err := extractPEM(stdout.String(), "CERTIFICATE")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(certDir, 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(keyPath, []byte(keyPEM), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(certPath, []byte(certPEM), 0o644); err != nil {
		return err
	}
	return nil
}

func extractPEM(raw, header string) (string, error) {
	begin := fmt.Sprintf("-----BEGIN %s-----", header)
	end := fmt.Sprintf("-----END %s-----", header)
	start := strings.Index(raw, begin)
	if start == -1 {
		return "", fmt.Errorf("unable to find %s in generated output", begin)
	}
	finish := strings.Index(raw[start:], end)
	if finish == -1 {
		return "", fmt.Errorf("unable to find %s terminator", end)
	}
	finish += start + len(end)
	block := raw[start:finish]
	return strings.TrimSpace(block) + "\n", nil
}

func writeCaddyFile(path string, data templates.Data) error {
	content, err := templates.RenderCaddy(data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(content, '\n'), 0o640)
}

func writeSubscription(dir, domain, body string) (string, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	target := filepath.Join(dir, fmt.Sprintf("%s.txt", domain))
	if err := os.WriteFile(target, []byte(body), 0o640); err != nil {
		return "", err
	}
	return target, nil
}
