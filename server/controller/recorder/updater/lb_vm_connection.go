/*
 * Copyright (c) 2024 KhulnaSoft, Ltd
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

type LBVMConnection struct {
	UpdaterBase[
		cloudmodel.LBVMConnection,
		*diffbase.LBVMConnection,
		*mysqlmodel.LBVMConnection,
		mysqlmodel.LBVMConnection,
		*message.LBVMConnectionAdd,
		message.LBVMConnectionAdd,
		*message.LBVMConnectionUpdate,
		message.LBVMConnectionUpdate,
		*message.LBVMConnectionFieldsUpdate,
		message.LBVMConnectionFieldsUpdate,
		*message.LBVMConnectionDelete,
		message.LBVMConnectionDelete]
}

func NewLBVMConnection(wholeCache *cache.Cache, cloudData []cloudmodel.LBVMConnection) *LBVMConnection {
	updater := &LBVMConnection{
		newUpdaterBase[
			cloudmodel.LBVMConnection,
			*diffbase.LBVMConnection,
			*mysqlmodel.LBVMConnection,
			mysqlmodel.LBVMConnection,
			*message.LBVMConnectionAdd,
			message.LBVMConnectionAdd,
			*message.LBVMConnectionUpdate,
			message.LBVMConnectionUpdate,
			*message.LBVMConnectionFieldsUpdate,
			message.LBVMConnectionFieldsUpdate,
			*message.LBVMConnectionDelete,
		](
			ctrlrcommon.RESOURCE_TYPE_LB_VM_CONNECTION_EN,
			wholeCache,
			db.NewLBVMConnection().SetMetadata(wholeCache.GetMetadata()),
			wholeCache.DiffBaseDataSet.LBVMConnections,
			cloudData,
		),
	}
	updater.dataGenerator = updater
	return updater
}

func (c *LBVMConnection) getDiffBaseByCloudItem(cloudItem *cloudmodel.LBVMConnection) (diffBase *diffbase.LBVMConnection, exists bool) {
	diffBase, exists = c.diffBaseData[cloudItem.Lcuuid]
	return
}

func (c *LBVMConnection) generateDBItemToAdd(cloudItem *cloudmodel.LBVMConnection) (*mysqlmodel.LBVMConnection, bool) {
	vmID, exists := c.cache.ToolDataSet.GetVMIDByLcuuid(cloudItem.VMLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			ctrlrcommon.RESOURCE_TYPE_VM_EN, cloudItem.VMLcuuid,
			ctrlrcommon.RESOURCE_TYPE_LB_VM_CONNECTION_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	lbID, exists := c.cache.ToolDataSet.GetLBIDByLcuuid(cloudItem.LBLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			ctrlrcommon.RESOURCE_TYPE_LB_EN, cloudItem.LBLcuuid,
			ctrlrcommon.RESOURCE_TYPE_LB_VM_CONNECTION_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}

	dbItem := &mysqlmodel.LBVMConnection{
		Domain: c.metadata.Domain.Lcuuid,
		VMID:   vmID,
		LBID:   lbID,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

// 保留接口
func (c *LBVMConnection) generateUpdateInfo(diffBase *diffbase.LBVMConnection, cloudItem *cloudmodel.LBVMConnection) (*message.LBVMConnectionFieldsUpdate, map[string]interface{}, bool) {
	return nil, nil, false
}
