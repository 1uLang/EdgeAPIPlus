// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/iwind/TeaGo/lists"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
)

type RequestKey struct {
	MacAddresses []string `json:"macAddresses"` // MAC 排序后内容
	MachineId    string   `json:"machineId"`    // /etc/machine-id
}

// GenerateRequestKey 生成请求Key
func GenerateRequestKey() (*RequestKey, error) {
	// mac addresses
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, errors.New("could not generate request key (code: 001)")
	}

	var macAddrs = []string{}
	for _, netInterface := range netInterfaces {
		if netInterface.Flags&net.FlagLoopback == net.FlagLoopback {
			continue
		}
		if netInterface.Flags&net.FlagUp == net.FlagUp {
			var macAddr = strings.TrimSpace(netInterface.HardwareAddr.String())
			if len(macAddr) == 0 {
				continue
			}
			macAddrs = append(macAddrs, macAddr)
		}
	}
	if len(macAddrs) == 0 {
		return nil, errors.New("could not generate request key (code: 002)")
	}
	sort.Strings(macAddrs)

	// machine id
	var machineId = ""
	var machineIdFile = "/etc/machine-id"
	stat, err := os.Stat(machineIdFile)
	if err == nil && !stat.IsDir() {
		data, err := os.ReadFile(machineIdFile)
		data = bytes.TrimSpace(data)
		if err == nil && len(data) <= 32 {
			machineId = string(data)
		}
	}

	return &RequestKey{
		MacAddresses: macAddrs,
		MachineId:    machineId,
	}, nil
}

// GenerateRequestCode 生成请求Key代码
func GenerateRequestCode() (string, error) {
	key, err := GenerateRequestKey()
	if err != nil {
		return "", err
	}
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return "", errors.New("could not generate request code (code: 001)")
	}
	return Encode(keyJSON)
}

// DecodeRequestCode 解析请求Key代码
func DecodeRequestCode(requestCode string) (*RequestKey, error) {
	requestCode = regexp.MustCompile(`\s+`).ReplaceAllString(requestCode, "")

	if requestCode == "*" {
		return &RequestKey{
			MacAddresses: nil,
			MachineId:    "",
		}, nil
	}
	m, err := DecodeData([]byte(requestCode))
	if err != nil {
		return nil, err
	}
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var key = &RequestKey{}
	err = json.Unmarshal(jsonData, key)
	return key, err
}

// ValidateRequestCode 校验请求Key代码
func ValidateRequestCode(requestCode string) (ok bool, errorCode string) {
	requestCode = regexp.MustCompile(`\s+`).ReplaceAllString(requestCode, "")

	if requestCode == "*" {
		return true, ""
	}

	key, err := DecodeRequestCode(requestCode)
	if err != nil {
		return false, "001"
	}

	// mac addresses
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return false, "002"
	}

	var allMACAddresses = []string{}
	for _, netInterface := range netInterfaces {
		var macAddr = strings.TrimSpace(netInterface.HardwareAddr.String())
		if len(macAddr) == 0 {
			continue
		}
		allMACAddresses = append(allMACAddresses, macAddr)
	}

	// check mac addresses
	for _, macAddress := range key.MacAddresses {
		if !lists.ContainsString(allMACAddresses, macAddress) {
			return false, "003"
		}
	}

	// check machine id
	if len(key.MachineId) > 0 {
		// machine id
		var machineId = ""
		var machineIdFile = "/etc/machine-id"
		stat, err := os.Stat(machineIdFile)
		if err == nil && !stat.IsDir() {
			data, err := os.ReadFile(machineIdFile)
			data = bytes.TrimSpace(data)
			if err == nil && len(data) <= 32 {
				machineId = string(data)
			}
		}
		if machineId != key.MachineId {
			return false, "004"
		}
	}

	return true, ""
}
