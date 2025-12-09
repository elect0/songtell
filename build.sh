#!/bin/bash

mkdir -p bin

echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o bin/songtell-linux-amd64

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o bin/songtell-windows-amd64.exe

echo "Building for macOS Intel/AMD..."
GOOS=darwin GOARCH=amd64 go build -o bin/songtell-darwin-amd64

echo "Building for macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -o bin/songtell-darwin-arm64

echo "--- Build complete. Files are in the 'bin/' directory. ---"
