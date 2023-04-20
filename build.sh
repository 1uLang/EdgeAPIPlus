#!/bin/bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags plus -o edge-api cmd/edge-api/main.go
