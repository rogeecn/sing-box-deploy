package share

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rogeecn/sing-box-deploy/internal/spec"
)

// BuildLink renders a subscription URL for supported protocols.
func BuildLink(inbound spec.InboundSpec, domain string) (string, error) {
	switch inbound.Protocol {
	case "vmess":
		return buildVMess(inbound, domain), nil
	case "vless":
		return buildVLESS(inbound, domain), nil
	default:
		return "", fmt.Errorf("share link for protocol %s is not supported", inbound.Protocol)
	}
}

func buildVMess(inbound spec.InboundSpec, domain string) string {
	payload := map[string]string{
		"v":    "2",
		"ps":   inbound.Tag,
		"add":  domain,
		"port": "443",
		"id":   inbound.UUID,
		"aid":  "0",
		"net":  transformTransport(inbound.Transport),
		"type": "none",
		"host": domain,
		"path": inbound.Path,
		"tls":  "tls",
	}
	raw, _ := json.Marshal(payload)
	encoded := base64.StdEncoding.EncodeToString(raw)
	return "vmess://" + encoded
}

func buildVLESS(inbound spec.InboundSpec, domain string) string {
	query := []string{
		"encryption=none",
		"security=tls",
		fmt.Sprintf("type=%s", transformTransport(inbound.Transport)),
		fmt.Sprintf("host=%s", domain),
		fmt.Sprintf("path=%s", inbound.Path),
	}
	return fmt.Sprintf(
		"vless://%s@%s:443?%s#%s",
		inbound.UUID,
		domain,
		strings.Join(query, "&"),
		inbound.Tag,
	)
}

func transformTransport(t string) string {
	switch strings.ToLower(t) {
	case "http":
		return "h2"
	case "ws":
		return "ws"
	default:
		return t
	}
}
