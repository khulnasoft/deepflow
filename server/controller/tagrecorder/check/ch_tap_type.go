/*
 * Copyright (c) 2023 KhulnaSoft, Ltd
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

package tagrecorder

import (
	"github.com/khulnasoft/deepflow/server/controller/db/mysql"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
)

type ChTapType struct {
	UpdaterBase[mysqlmodel.ChTapType, TapTypeKey]
	resourceTypeToIconID map[IconKey]int
}

func NewChTapType(resourceTypeToIconID map[IconKey]int) *ChTapType {
	updater := &ChTapType{
		UpdaterBase[mysqlmodel.ChTapType, TapTypeKey]{
			resourceTypeName: RESOURCE_TYPE_TAP_TYPE,
		},
		resourceTypeToIconID,
	}
	updater.dataGenerator = updater
	return updater
}

func (t *ChTapType) generateNewData() (map[TapTypeKey]mysqlmodel.ChTapType, bool) {
	var tapTypes []mysqlmodel.TapType
	err := mysql.DefaultDB.Unscoped().Find(&tapTypes).Error
	if err != nil {
		log.Errorf(dbQueryResourceFailed(t.resourceTypeName, err), t.db.LogPrefixORGID)
		return nil, false
	}

	keyToItem := make(map[TapTypeKey]mysqlmodel.ChTapType)
	for _, tapType := range tapTypes {
		keyToItem[TapTypeKey{Value: tapType.Value}] = mysqlmodel.ChTapType{
			Value: tapType.Value,
			Name:  tapType.Name,
		}
	}
	return keyToItem, true
}

func (t *ChTapType) generateKey(dbItem mysqlmodel.ChTapType) TapTypeKey {
	return TapTypeKey{Value: dbItem.Value}
}

func (t *ChTapType) generateUpdateInfo(oldItem, newItem mysqlmodel.ChTapType) (map[string]interface{}, bool) {
	updateInfo := make(map[string]interface{})
	if oldItem.Name != newItem.Name {
		updateInfo["name"] = newItem.Name
	}
	if len(updateInfo) > 0 {
		return updateInfo, true
	}
	return nil, false
}
