package spec

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// InboundSpec describes a single inbound entry rendered through templates.
type InboundSpec struct {
	Key        string `json:"key"`
	Tag        string `json:"tag"`
	FileName   string `json:"file_name"`
	Protocol   string `json:"protocol"`
	Listen     string `json:"listen"`
	ListenPort int    `json:"listen_port"`
	UUID       string `json:"uuid"`
	Password   string `json:"password,omitempty"`
	Path       string `json:"path"`
	Host       string `json:"host"`
	Transport  string `json:"transport"`
}

type definition struct {
	Protocol  string
	Transport string
	TagFormat string
}

var definitions = map[string]definition{
	"vless-h2-tls": {
		Protocol:  "vless",
		Transport: "http",
		TagFormat: "VLESS-H2-TLS-%s.json",
	},
	"vless-httpupgrade-tls": {
		Protocol:  "vless",
		Transport: "httpupgrade",
		TagFormat: "VLESS-HTTPUpgrade-TLS-%s.json",
	},
	"vless-ws-tls": {
		Protocol:  "vless",
		Transport: "ws",
		TagFormat: "VLESS-WS-TLS-%s.json",
	},
	"vmess-h2-tls": {
		Protocol:  "vmess",
		Transport: "http",
		TagFormat: "VMess-H2-TLS-%s.json",
	},
	"vmess-httpupgrade-tls": {
		Protocol:  "vmess",
		Transport: "httpupgrade",
		TagFormat: "VMess-HTTPUpgrade-TLS-%s.json",
	},
	"vmess-ws-tls": {
		Protocol:  "vmess",
		Transport: "ws",
		TagFormat: "VMess-WS-TLS-%s.json",
	},
}

// SupportedKeys returns the inbound identifiers supported by templates.
func SupportedKeys() []string {
	keys := make([]string, 0, len(definitions))
	for key := range definitions {
		keys = append(keys, key)
	}
	return keys
}

// Exists reports whether the inbound key is defined.
func Exists(key string) bool {
	_, ok := definitions[key]
	return ok
}

// BuildSpec generates a spec pre-populated with default values for the domain.
func BuildSpec(key, domain string) (InboundSpec, error) {
	def, ok := definitions[key]
	if !ok {
		return InboundSpec{}, fmt.Errorf("unsupported inbound type: %s", key)
	}
	uid := newUUID()
	tag := fmt.Sprintf(def.TagFormat, domain)
	port, err := randomHighPort()
	if err != nil {
		return InboundSpec{}, err
	}
	return InboundSpec{
		Key:        key,
		Tag:        tag,
		FileName:   tag,
		Protocol:   def.Protocol,
		Listen:     "127.0.0.1",
		ListenPort: port,
		UUID:       uid,
		Path:       "/" + uid,
		Host:       domain,
		Transport:  def.Transport,
	}, nil
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		binary.BigEndian.Uint32(b[0:4]),
		binary.BigEndian.Uint16(b[4:6]),
		binary.BigEndian.Uint16(b[6:8]),
		binary.BigEndian.Uint16(b[8:10]),
		binary.BigEndian.Uint16(b[10:12]),
		binary.BigEndian.Uint32(b[12:16]),
	)
}

func randomHighPort() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(math.MaxUint16-32768)))
	if err != nil {
		return 0, fmt.Errorf("random port: %w", err)
	}
	return 32768 + int(n.Int64()), nil
}

// NormalizeKeys deduplicates inbound keys and preserves order.
func NormalizeKeys(keys []string) ([]string, error) {
	if len(keys) == 0 {
		return SupportedKeys(), nil
	}
	seen := make(map[string]struct{})
	var normalized []string
	for _, key := range keys {
		k := strings.TrimSpace(strings.ToLower(key))
		if k == "" {
			continue
		}
		if _, ok := definitions[k]; !ok {
			return nil, fmt.Errorf("unknown inbound type %q", key)
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		normalized = append(normalized, k)
	}
	return normalized, nil
}
