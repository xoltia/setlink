package web

import "embed"

//go:embed template/*
var Templates embed.FS

//go:embed static/*
var StaticAssets embed.FS
