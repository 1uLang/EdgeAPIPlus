// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	teaconst "github.com/TeaOSLab/EdgePlus/pkg/const"
	"github.com/TeaOSLab/EdgePlus/pkg/encrypt"
	"github.com/iwind/TeaGo/maps"
	"time"
)

// Encode 加密
func Encode(data []byte) (string, error) {
	instance, err := encrypt.NewMethodInstance("aes-256-cfb", teaconst.PlusKey, teaconst.PlusIV)
	if err != nil {
		return "", errors.New("不支持选择的加密方式")
	}
	dist, err := instance.Encrypt(data)
	if err != nil {
		return "", errors.New("加密失败：" + err.Error())
	}
	return base64.StdEncoding.EncodeToString(dist), nil
}

// EncodeMap 加密Map
func EncodeMap(m maps.Map) (string, error) {
	m["updatedAt"] = time.Now().Unix() // 用来校验Authority服务是否已经更新

	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return Encode(data)
}

// DecodeData 解密
func DecodeData(data []byte) (maps.Map, error) {
	instance, err := encrypt.NewMethodInstance("aes-256-cfb", teaconst.PlusKey, teaconst.PlusIV)
	if err != nil {
		return nil, errors.New("encrypt method not supported")
	}
	source, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(data)))
	if err != nil {
		return nil, errors.New("decode key failed: base64 decode failed: " + err.Error())
	}
	dist, err := instance.Decrypt(source)
	if err != nil {
		return nil, errors.New("decode key failed: decrypt failed: " + err.Error())
	}
	var m = maps.Map{}
	err = json.Unmarshal(dist, &m)
	if err != nil {
		return nil, errors.New("decode key failed: decode json failed: " + err.Error())
	}

	return m, nil
}

func Decode(data []byte) (maps.Map, error) {
	m, err := DecodeData(data)
	if err != nil {
		return nil, err
	}

	// 控制 STILL 用户权限
	if m.GetString("company") == "STILL" {
		m["components"] = []ComponentCode{
			ComponentCodeLog,
			ComponentCodeNS,
			ComponentCodeUser,
		}
	}

	if len(m.GetString("dayFrom")) == 0 || len(m.GetString("dayTo")) == 0 || m.GetInt("nodes") <= 0 {
		return nil, errors.New("invalid key")
	}
	return m, nil
}

// EncodeKey 加密Key
func EncodeKey(key *Key) (string, error) {
	key.UpdatedAt = time.Now().Unix() // 用来校验Authority服务是否已经更新
	data, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	return Encode(data)
}

// DecodeKey 解密Key
func DecodeKey(data []byte) (*Key, error) {
	instance, err := encrypt.NewMethodInstance("aes-256-cfb", teaconst.PlusKey, teaconst.PlusIV)
	if err != nil {
		return nil, errors.New("encrypt method not supported")
	}
	source, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(data)))
	if err != nil {
		return nil, errors.New("decode key failed: base64 decode failed: " + err.Error())
	}
	dist, err := instance.Decrypt(source)
	if err != nil {
		return nil, errors.New("decode key failed: decrypt failed: " + err.Error())
	}

	var result = &Key{}
	err = json.Unmarshal(dist, result)
	if err != nil {
		return nil, errors.New("decode key failed: " + err.Error())
	}

	// 这里不能限制节点，因为以往有不限节点的授权
	if len(result.DayFrom) == 0 || len(result.DayTo) == 0 {
		return nil, errors.New("invalid key")
	}

	// 控制 STILL 用户权限
	if result.Company == "STILL" {
		result.Components = []ComponentCode{
			ComponentCodeLog,
			ComponentCodeNS,
			ComponentCodeUser,
		}
	}

	return result, nil
}
