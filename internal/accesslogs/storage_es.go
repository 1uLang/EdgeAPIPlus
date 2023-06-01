//go:build plus

package accesslogs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/maps"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ESStorage ElasticSearch存储策略
type ESStorage struct {
	BaseStorage

	config *serverconfigs.AccessLogESStorageConfig
}

func NewESStorage(config *serverconfigs.AccessLogESStorageConfig) *ESStorage {
	return &ESStorage{config: config}
}

func (this *ESStorage) Config() any {
	return this.config
}

// Start 开启
func (this *ESStorage) Start() error {
	if len(this.config.Endpoint) == 0 {
		return errors.New("'endpoint' should not be nil")
	}
	if !regexp.MustCompile(`(?i)^(http|https)://`).MatchString(this.config.Endpoint) {
		this.config.Endpoint = "http://" + this.config.Endpoint
	}

	// 去除endpoint中的路径部分
	u, err := url.Parse(this.config.Endpoint)
	if err == nil && len(u.Path) > 0 {
		this.config.Endpoint = u.Scheme + "://" + u.Host
	}

	if len(this.config.Index) == 0 {
		return errors.New("'index' should not be nil")
	}
	if !this.config.IsDataStream && len(this.config.MappingType) == 0 {
		return errors.New("'mappingType' should not be nil")
	}

	return nil
}

// 写入日志
func (this *ESStorage) Write(accessLogs []*pb.HTTPAccessLog) error {
	if len(accessLogs) == 0 {
		return nil
	}

	var bulk = &strings.Builder{}
	var indexName = this.FormatVariables(this.config.Index)
	var typeName = this.FormatVariables(this.config.MappingType)
	for _, accessLog := range accessLogs {
		if this.firewallOnly && accessLog.FirewallPolicyId == 0 {
			continue
		}

		if len(accessLog.RequestId) == 0 {
			continue
		}

		var indexMap = map[string]any{
			"_index": indexName,
			"_id":    accessLog.RequestId,
		}
		if !this.config.IsDataStream {
			indexMap["_type"] = typeName
		}
		opData, err := json.Marshal(map[string]any{
			"index": indexMap,
		})
		if err != nil {
			remotelogs.Error("ACCESS_LOG_ES_STORAGE", "write failed: "+err.Error())
			continue
		}

		data, err := this.Marshal(accessLog)
		if err != nil {
			remotelogs.Error("ACCESS_LOG_ES_STORAGE", "marshal data failed: "+err.Error())
			continue
		}

		bulk.Write(opData)
		bulk.WriteString("\n")
		bulk.Write(data)
		bulk.WriteString("\n")
	}

	if bulk.Len() == 0 {
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, this.config.Endpoint+"/_bulk", strings.NewReader(bulk.String()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", strings.ReplaceAll(teaconst.ProductName, " ", "-")+"/"+teaconst.Version)
	if len(this.config.Username) > 0 || len(this.config.Password) > 0 {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(this.config.Username+":"+this.config.Password)))
	}
	var client = utils.SharedHttpClient(10 * time.Second)
	defer func() {
		_ = req.Body.Close()
	}()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyData, _ := io.ReadAll(resp.Body)
		return errors.New("ElasticSearch response status code: " + fmt.Sprintf("%d", resp.StatusCode) + " content: " + string(bodyData))
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("read ElasticSearch response failed: " + err.Error())
	}
	var m = maps.Map{}
	err = json.Unmarshal(bodyData, &m)
	if err == nil {
		// 暂不处理非JSON的情况
		if m.Has("errors") && m.GetBool("errors") {
			return errors.New("ElasticSearch returns '" + string(bodyData) + "'")
		}
	}

	return nil
}

// Close 关闭
func (this *ESStorage) Close() error {
	return nil
}
