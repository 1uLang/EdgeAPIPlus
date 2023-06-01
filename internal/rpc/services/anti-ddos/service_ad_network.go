// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build plus

package antiddos

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type ADNetworkService struct {
	services.BaseService
}

// CreateADNetwork 创建线路
func (this *ADNetworkService) CreateADNetwork(ctx context.Context, req *pb.CreateADNetworkRequest) (*pb.CreateADNetworkResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	networkId, err := models.SharedADNetworkDAO.CreateNetwork(tx, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateADNetworkResponse{AdNetworkId: networkId}, nil
}

// UpdateADNetwork 修改线路
func (this *ADNetworkService) UpdateADNetwork(ctx context.Context, req *pb.UpdateADNetworkRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if req.AdNetworkId <= 0 {
		return nil, errors.New("invalid adNetworkId")
	}
	err = models.SharedADNetworkDAO.UpdateNetwork(tx, req.AdNetworkId, req.IsOn, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindADNetwork 查找单个线路
func (this *ADNetworkService) FindADNetwork(ctx context.Context, req *pb.FindADNetworkRequest) (*pb.FindADNetworkResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	network, err := models.SharedADNetworkDAO.FindEnabledADNetwork(tx, req.AdNetworkId)
	if err != nil {
		return nil, err
	}
	if network == nil {
		return &pb.FindADNetworkResponse{AdNetwork: nil}, nil
	}

	return &pb.FindADNetworkResponse{AdNetwork: &pb.ADNetwork{
		Id:          int64(network.Id),
		IsOn:        network.IsOn,
		Name:        network.Name,
		Description: network.Description,
	}}, nil
}

// FindAllADNetworks 列出所有线路
func (this *ADNetworkService) FindAllADNetworks(ctx context.Context, req *pb.FindAllADNetworkRequest) (*pb.FindAllADNetworkResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	networks, err := models.SharedADNetworkDAO.FindAllNetworks(tx)
	if err != nil {
		return nil, err
	}
	var pbNetworks = []*pb.ADNetwork{}
	for _, network := range networks {
		pbNetworks = append(pbNetworks, &pb.ADNetwork{
			Id:          int64(network.Id),
			IsOn:        network.IsOn,
			Name:        network.Name,
			Description: network.Description,
		})
	}
	return &pb.FindAllADNetworkResponse{
		AdNetworks: pbNetworks,
	}, nil
}

// FindAllAvailableADNetworks 列出所有可用的线路
func (this *ADNetworkService) FindAllAvailableADNetworks(ctx context.Context, req *pb.FindAllAvailableADNetworksRequest) (*pb.FindAllAvailableADNetworksResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	networks, err := models.SharedADNetworkDAO.FindAllAvailableNetworks(tx)
	if err != nil {
		return nil, err
	}
	var pbNetworks = []*pb.ADNetwork{}
	for _, network := range networks {
		pbNetworks = append(pbNetworks, &pb.ADNetwork{
			Id:          int64(network.Id),
			IsOn:        network.IsOn,
			Name:        network.Name,
			Description: network.Description,
		})
	}
	return &pb.FindAllAvailableADNetworksResponse{
		AdNetworks: pbNetworks,
	}, nil
}

// DeleteADNetwork 删除线路
func (this *ADNetworkService) DeleteADNetwork(ctx context.Context, req *pb.DeleteADNetworkRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedADNetworkDAO.DisableADNetwork(tx, req.AdNetworkId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
