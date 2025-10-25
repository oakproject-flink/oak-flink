//go:build tools
// +build tools

// Package tools tracks build tool dependencies in go.mod
// This is a Go best practice for tracking development tools
// See: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)