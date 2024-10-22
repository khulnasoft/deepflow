/*
 * Copyright (c) 2024 Yunshan Networks
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package updater

import (
	cloudmodel "github.com/khulnasoft/deepflow/server/controller/cloud/model"
	ctrlrcommon "github.com/khulnasoft/deepflow/server/controller/common"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
	"github.com/khulnasoft/deepflow/server/controller/recorder/cache"
	"github.com/khulnasoft/deepflow/server/controller/recorder/cache/diffbase"
	"github.com/khulnasoft/deepflow/server/controller/recorder/db"
	"github.com/khulnasoft/deepflow/server/controller/recorder/pubsub/message"
)

type PeerConnection struct {
	UpdaterBase[
		cloudmodel.PeerConnection,
		*diffbase.PeerConnection,
		*mysqlmodel.PeerConnection,
		mysqlmodel.PeerConnection,
		*message.PeerConnectionAdd,
		message.PeerConnectionAdd,
		*message.PeerConnectionUpdate,
		message.PeerConnectionUpdate,
		*message.PeerConnectionFieldsUpdate,
		message.PeerConnectionFieldsUpdate,
		*message.PeerConnectionDelete,
		message.PeerConnectionDelete]
}

func NewPeerConnection(wholeCache *cache.Cache, cloudData []cloudmodel.PeerConnection) *PeerConnection {
	updater := &PeerConnection{
		newUpdaterBase[
			cloudmodel.PeerConnection,
			*diffbase.PeerConnection,
			*mysqlmodel.PeerConnection,
			mysqlmodel.PeerConnection,
			*message.PeerConnectionAdd,
			message.PeerConnectionAdd,
			*message.PeerConnectionUpdate,
			message.PeerConnectionUpdate,
			*message.PeerConnectionFieldsUpdate,
			message.PeerConnectionFieldsUpdate,
			*message.PeerConnectionDelete,
		](
			ctrlrcommon.RESOURCE_TYPE_PEER_CONNECTION_EN,
			wholeCache,
			db.NewPeerConnection().SetMetadata(wholeCache.GetMetadata()),
			wholeCache.DiffBaseDataSet.PeerConnections,
			cloudData,
		),
	}
	updater.dataGenerator = updater
	return updater
}

func (c *PeerConnection) getDiffBaseByCloudItem(cloudItem *cloudmodel.PeerConnection) (diffBase *diffbase.PeerConnection, exists bool) {
	diffBase, exists = c.diffBaseData[cloudItem.Lcuuid]
	return
}

func (c *PeerConnection) generateDBItemToAdd(cloudItem *cloudmodel.PeerConnection) (*mysqlmodel.PeerConnection, bool) {
	remoteVPCID, exists := c.cache.ToolDataSet.GetVPCIDByLcuuid(cloudItem.RemoteVPCLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			ctrlrcommon.RESOURCE_TYPE_VPC_EN, cloudItem.RemoteVPCLcuuid,
			ctrlrcommon.RESOURCE_TYPE_PEER_CONNECTION_EN, cloudItem.Lcuuid,
		), c.metadata.LogPrefixes)
		return nil, false
	}
	localVPCID, exists := c.cache.ToolDataSet.GetVPCIDByLcuuid(cloudItem.LocalVPCLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			ctrlrcommon.RESOURCE_TYPE_VPC_EN, cloudItem.LocalVPCLcuuid,
			ctrlrcommon.RESOURCE_TYPE_PEER_CONNECTION_EN, cloudItem.Lcuuid,
		), c.metadata.LogPrefixes)
		return nil, false
	}
	remoteRegionID, exists := c.cache.ToolDataSet.GetRegionIDByLcuuid(cloudItem.RemoteRegionLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			ctrlrcommon.RESOURCE_TYPE_REGION_EN, cloudItem.RemoteRegionLcuuid,
			ctrlrcommon.RESOURCE_TYPE_PEER_CONNECTION_EN, cloudItem.Lcuuid,
		), c.metadata.LogPrefixes)
		return nil, false
	}
	localRegionID, exists := c.cache.ToolDataSet.GetRegionIDByLcuuid(cloudItem.LocalRegionLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			ctrlrcommon.RESOURCE_TYPE_REGION_EN, cloudItem.LocalRegionLcuuid,
			ctrlrcommon.RESOURCE_TYPE_PEER_CONNECTION_EN, cloudItem.Lcuuid,
		), c.metadata.LogPrefixes)
		return nil, false
	}
	dbItem := &mysqlmodel.PeerConnection{
		Name:           cloudItem.Name,
		Label:          cloudItem.Label,
		Domain:         c.metadata.Domain.Lcuuid,
		RemoteVPCID:    remoteVPCID,
		LocalVPCID:     localVPCID,
		RemoteRegionID: remoteRegionID,
		LocalRegionID:  localRegionID,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

func (c *PeerConnection) generateUpdateInfo(diffBase *diffbase.PeerConnection, cloudItem *cloudmodel.PeerConnection) (*message.PeerConnectionFieldsUpdate, map[string]interface{}, bool) {
	structInfo := new(message.PeerConnectionFieldsUpdate)
	mapInfo := make(map[string]interface{})
	if diffBase.Name != cloudItem.Name {
		mapInfo["name"] = cloudItem.Name
		structInfo.Name.Set(diffBase.Name, cloudItem.Name)
	}
	if diffBase.RemoteRegionLcuuid != cloudItem.RemoteRegionLcuuid {
		remoteRegionID, exists := c.cache.ToolDataSet.GetRegionIDByLcuuid(cloudItem.RemoteRegionLcuuid)
		if !exists {
			log.Error(resourceAForResourceBNotFound(
				ctrlrcommon.RESOURCE_TYPE_REGION_EN, cloudItem.RemoteRegionLcuuid,
				ctrlrcommon.RESOURCE_TYPE_PEER_CONNECTION_EN, cloudItem.Lcuuid,
			), c.metadata.LogPrefixes)
			return nil, nil, false
		}
		mapInfo["remote_region_id"] = remoteRegionID
		structInfo.RemoteRegionID.SetNew(remoteRegionID)
		structInfo.RemoteRegionLcuuid.Set(diffBase.RemoteRegionLcuuid, cloudItem.RemoteRegionLcuuid)
	}
	if diffBase.LocalRegionLcuuid != cloudItem.LocalRegionLcuuid {
		localRegionID, exists := c.cache.ToolDataSet.GetRegionIDByLcuuid(cloudItem.LocalRegionLcuuid)
		if !exists {
			log.Error(resourceAForResourceBNotFound(
				ctrlrcommon.RESOURCE_TYPE_REGION_EN, cloudItem.LocalRegionLcuuid,
				ctrlrcommon.RESOURCE_TYPE_PEER_CONNECTION_EN, cloudItem.Lcuuid,
			), c.metadata.LogPrefixes)
			return nil, nil, false
		}
		mapInfo["local_region_id"] = localRegionID
		structInfo.LocalRegionID.SetNew(localRegionID)
		structInfo.LocalRegionLcuuid.Set(diffBase.LocalRegionLcuuid, cloudItem.LocalRegionLcuuid)
	}

	return structInfo, mapInfo, len(mapInfo) > 0
}
