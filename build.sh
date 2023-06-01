#!/bin/bash
go mod tidy 
go build -tags plus -o edge-api cmd/edge-api/main.go 
cp edge-api ../EdgeAdmin/build/edge-api/bin 
