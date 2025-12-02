package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var ErrNotFound = errors.New("state file not found")

type Inbound struct {
	Key        string `json:"key"`
	Tag        string `json:"tag"`
	Protocol   string `json:"protocol"`
	Transport  string `json:"transport"`
	ListenPort int    `json:"listen_port"`
	UUID       string `json:"uuid"`
	Path       string `json:"path"`
	Host       string `json:"host"`
	ShareURL   string `json:"share_url"`
}

type State struct {
	Domain           string    `json:"domain"`
	Email            string    `json:"email"`
	RootDir          string    `json:"root_dir"`
	CaddyFile        string    `json:"caddy_file"`
	SubscriptionFile string    `json:"subscription_file"`
	Inbounds         []Inbound `json:"inbounds"`
	LastUpdated      time.Time `json:"last_updated"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	return &st, nil
}

func Save(path string, st *State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	st.LastUpdated = time.Now().UTC()
	payload, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	return os.WriteFile(path, payload, 0o644)
}
