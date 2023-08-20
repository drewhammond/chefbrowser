package ui

import "embed"

// Embedded contains embedded UI resources
//
//go:embed templates/* dist/*
var Embedded embed.FS
