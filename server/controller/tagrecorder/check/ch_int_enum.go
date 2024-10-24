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
	"strconv"
	"strings"

	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
	"github.com/khulnasoft/deepflow/server/querier/config"
	"github.com/khulnasoft/deepflow/server/querier/engine/clickhouse/tag"
)

type ChIntEnum struct {
	UpdaterBase[mysqlmodel.ChIntEnum, IntEnumTagKey]
}

func NewChIntEnum() *ChIntEnum {
	updater := &ChIntEnum{
		UpdaterBase[mysqlmodel.ChIntEnum, IntEnumTagKey]{
			resourceTypeName: RESOURCE_TYPE_CH_INT_ENUM,
		},
	}
	updater.dataGenerator = updater
	return updater
}

func (e *ChIntEnum) generateNewData() (map[IntEnumTagKey]mysqlmodel.ChIntEnum, bool) {
	sql := "show tag all_int_enum values from tagrecorder"
	db := "tagrecorder"
	table := "tagrecorder"
	keyToItem := make(map[IntEnumTagKey]mysqlmodel.ChIntEnum)
	respMap, err := tag.GetEnumTagValues(db, table, sql)
	if err != nil {
		log.Errorf("read failed: %v", err, e.db.LogPrefixORGID)
	}

	for name, tagValues := range respMap {
		tagName := strings.TrimSuffix(name, "."+config.Cfg.Language)
		for _, valueAndName := range tagValues {
			tagValue := valueAndName.([]interface{})[0]
			tagDisplayName := valueAndName.([]interface{})[1]
			tagDescription := valueAndName.([]interface{})[2]
			tagValueInt, err := strconv.Atoi(tagValue.(string))
			if err == nil {
				key := IntEnumTagKey{
					TagName:  tagName,
					TagValue: tagValueInt,
				}
				keyToItem[key] = mysqlmodel.ChIntEnum{
					TagName:     tagName,
					Value:       tagValueInt,
					Name:        tagDisplayName.(string),
					Description: tagDescription.(string),
				}
			}

		}
	}

	return keyToItem, true
}

func (e *ChIntEnum) generateKey(dbItem mysqlmodel.ChIntEnum) IntEnumTagKey {
	return IntEnumTagKey{TagName: dbItem.TagName, TagValue: dbItem.Value}
}

func (e *ChIntEnum) generateUpdateInfo(oldItem, newItem mysqlmodel.ChIntEnum) (map[string]interface{}, bool) {
	updateInfo := make(map[string]interface{})
	if oldItem.TagName != newItem.TagName {
		updateInfo["tag_name"] = newItem.TagName
	}
	if oldItem.Value != newItem.Value {
		updateInfo["value"] = newItem.Value
	}
	if oldItem.Name != newItem.Name {
		updateInfo["name"] = newItem.Name
	}
	if oldItem.Description != newItem.Description {
		updateInfo["description"] = newItem.Description
	}
	if len(updateInfo) > 0 {
		return updateInfo, true
	}
	return nil, false
}
