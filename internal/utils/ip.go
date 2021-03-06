package utils

import (
	"encoding/binary"
	"github.com/cespare/xxhash/v2"
	"math"
	"net"
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
