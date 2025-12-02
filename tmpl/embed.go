package tmpl

import "embed"

// Files exposes embedded template assets.
//
//go:embed **/*
var Files embed.FS
