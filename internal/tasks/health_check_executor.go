package tasks

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/1uLang/EdgeCommon/pkg/configutils"
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/1uLang/EdgeCommon/pkg/nodeutils"
	"github.com/1uLang/EdgeCommon/pkg/serverconfigs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HealthCheckExecutor struct {
	clusterId int64
}

func NewHealthCheckExecutor(clusterId int64) *HealthCheckExecutor {
	return &HealthCheckExecutor{clusterId: clusterId}
}

func (this *HealthCheckExecutor) Run() ([]*HealthCheckResult, error) {
	cluster, err := models.NewNodeClusterDAO().FindEnabledNodeCluster(nil, this.clusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New("can not find cluster with id '" + strconv.FormatInt(this.clusterId, 10) + "'")
	}
	if len(cluster.HealthCheck) == 0 || cluster.HealthCheck == "null" {
		return nil, errors.New("health check config is not found")
	}

	healthCheckConfig := &serverconfigs.HealthCheckConfig{}
	err = json.Unmarshal([]byte(cluster.HealthCheck), healthCheckConfig)
	if err != nil {
		return nil, err
	}

	results := []*HealthCheckResult{}
	nodes, err := models.NewNodeDAO().FindAllEnabledNodesWithClusterId(nil, this.clusterId)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if node.IsOn != 1 {
			continue
		}
		result := &HealthCheckResult{
			Node: node,
		}

		ipAddr, ipAddrId, err := models.NewNodeIPAddressDAO().FindFirstNodeAccessIPAddress(nil, int64(node.Id), nodeconfigs.NodeRoleNode)
		if err != nil {
			return nil, err
		}
		if len(ipAddr) == 0 {
			result.Error = "no ip address can be used"
		} else {
			result.NodeAddr = ipAddr
			result.NodeAddrId = ipAddrId
		}

		results = append(results, result)
	}

	// ????????????
	preparedResults := []*HealthCheckResult{}
	for _, result := range results {
		if len(result.NodeAddr) > 0 {
			preparedResults = append(preparedResults, result)
		}
	}

	if len(preparedResults) == 0 {
		return results, nil
	}

	countResults := len(preparedResults)
	queue := make(chan *HealthCheckResult, countResults)
	for _, result := range preparedResults {
		queue <- result
	}

	countTries := types.Int(healthCheckConfig.CountTries)
	if countTries > 10 { // ??????????????????10??? TODO ?????????????????????????????????
		countTries = 10
	}
	if countTries < 1 {
		countTries = 3
	}

	tryDelay := 1 * time.Second
	if healthCheckConfig.TryDelay != nil {
		tryDelay = healthCheckConfig.TryDelay.Duration()

		if tryDelay > 1*time.Minute { // ??????????????????1?????? TODO ?????????????????????????????????
			tryDelay = 1 * time.Minute
		}
	}

	countRoutines := 10
	wg := sync.WaitGroup{}
	wg.Add(countResults)
	for i := 0; i < countRoutines; i++ {
		go func() {
			for {
				select {
				case result := <-queue:
					func() {
						for i := 1; i <= countTries; i++ {
							before := time.Now()
							err := this.checkNode(healthCheckConfig, result)
							result.CostMs = time.Since(before).Seconds() * 1000
							if err != nil {
								result.Error = err.Error()
							}
							if result.IsOk {
								break
							}
							if tryDelay > 0 {
								time.Sleep(tryDelay)
							}
						}

						// ????????????IP??????
						if teaconst.IsPlus {
							isChanged, err := models.SharedNodeIPAddressDAO.UpdateAddressHealthCount(nil, result.NodeAddrId, result.IsOk, healthCheckConfig.CountUp, healthCheckConfig.CountDown)
							if err != nil {
								remotelogs.Error("HEALTH_CHECK_EXECUTOR", err.Error())
								return
							}

							if isChanged {
								// ????????????
								err = models.SharedNodeIPAddressDAO.FireThresholds(nil, nodeconfigs.NodeRoleNode, int64(result.Node.Id))
								if err != nil {
									remotelogs.Error("HEALTH_CHECK_EXECUTOR", err.Error())
									return
								}
							}
						}

						// ??????????????????
						if healthCheckConfig.AutoDown {
							isChanged, err := models.SharedNodeDAO.UpdateNodeUpCount(nil, int64(result.Node.Id), result.IsOk, healthCheckConfig.CountUp, healthCheckConfig.CountDown)
							if err != nil {
								remotelogs.Error("HEALTH_CHECK_EXECUTOR", err.Error())
							} else if isChanged {
								// ?????????????????????
								if result.IsOk {
									message := "???????????????????????????\"" + result.Node.Name + "\"???????????????"
									err = models.NewMessageDAO().CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, int64(result.Node.Id), models.MessageTypeHealthCheckNodeUp, models.MessageLevelSuccess, message, message, nil, false)
								} else {
									message := "???????????????????????????\"" + result.Node.Name + "\"???????????????"
									err = models.NewMessageDAO().CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, int64(result.Node.Id), models.MessageTypeHealthCheckNodeDown, models.MessageLevelError, message, message, nil, false)
								}
							}
						}
					}()

					wg.Done()
				default:
					return
				}
			}
		}()
	}
	wg.Wait()

	return results, nil
}

// ??????????????????
func (this *HealthCheckExecutor) checkNode(healthCheckConfig *serverconfigs.HealthCheckConfig, result *HealthCheckResult) error {
	// ??????IPv6
	if utils.IsIPv6(result.NodeAddr) {
		result.NodeAddr = "[" + result.NodeAddr + "]"
	}

	if len(healthCheckConfig.URL) == 0 {
		healthCheckConfig.URL = "http://${host}/"
	}

	url := strings.ReplaceAll(healthCheckConfig.URL, "${host}", result.NodeAddr)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if len(healthCheckConfig.UserAgent) > 0 {
		req.Header.Set("User-Agent", healthCheckConfig.UserAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	}

	key, err := nodeutils.EncryptData(result.Node.UniqueId, result.Node.Secret, maps.Map{
		"onlyBasicRequest": healthCheckConfig.OnlyBasicRequest,
	}, 300)
	if err != nil {
		return err
	}
	req.Header.Set(serverconfigs.HealthCheckHeaderName, key)

	timeout := 5 * time.Second
	if healthCheckConfig.Timeout != nil {
		timeout = healthCheckConfig.Timeout.Duration()
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				_, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				conn, err := net.Dial(network, configutils.QuoteIP(result.NodeAddr)+":"+port)
				if err == nil {
					return conn, nil
				}
				return net.DialTimeout(network, configutils.QuoteIP(result.NodeAddr)+":"+port, timeout)
			},
			MaxIdleConns:          1,
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       1,
			IdleConnTimeout:       2 * time.Minute,
			ExpectContinueTimeout: 1 * time.Second,
			TLSHandshakeTimeout:   0,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	defer func() {
		client.CloseIdleConnections()
	}()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if len(healthCheckConfig.StatusCodes) > 0 && !lists.ContainsInt(healthCheckConfig.StatusCodes, resp.StatusCode) {
		result.Error = "invalid response status code '" + strconv.Itoa(resp.StatusCode) + "'"
		return nil
	}

	result.IsOk = true

	return nil
}
