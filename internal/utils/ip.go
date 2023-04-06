package utils

import (
	"encoding/binary"
	"github.com/cespare/xxhash/v2"
	"math"
	"net"
	"regexp"
	"strings"
)

// IP2Long 将IP转换为整型
// 注意IPv6没有顺序
func IP2Long(ip string) uint64 {
	if len(ip) == 0 {
		return 0
	}
	s := net.ParseIP(ip)
	if len(s) == 0 {
		return 0
	}

	if strings.Contains(ip, ":") {
		return math.MaxUint32 + xxhash.Sum64(s)
	}
	return uint64(binary.BigEndian.Uint32(s.To4()))
}

// IsIPv6 判断是否为IPv6
func IsIPv6(ip string) bool {
	return strings.Contains(ip, ":")
}

// 判断是否为IPv4
func IsIPv4(ip string) bool {
	if !regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`).MatchString(ip) {
		return false
	}
	if IP2Long(ip) == 0 {
		return false
	}
	return true
}
