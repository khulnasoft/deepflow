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

package tagrecorder

import (
	"encoding/json"

	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/controller/db/mysql"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
	"github.com/khulnasoft/deepflow/server/controller/recorder/pubsub/message"
	"github.com/khulnasoft/deepflow/server/libs/logger"
)

type ChPodNSCloudTags struct {
	SubscriberComponent[*message.PodNamespaceFieldsUpdate, message.PodNamespaceFieldsUpdate, mysqlmodel.PodNamespace, mysqlmodel.ChPodNSCloudTags, CloudTagsKey]
}

func NewChPodNSCloudTags() *ChPodNSCloudTags {
	mng := &ChPodNSCloudTags{
		newSubscriberComponent[*message.PodNamespaceFieldsUpdate, message.PodNamespaceFieldsUpdate, mysqlmodel.PodNamespace, mysqlmodel.ChPodNSCloudTags, CloudTagsKey](
			common.RESOURCE_TYPE_POD_NAMESPACE_EN, RESOURCE_TYPE_CH_POD_NS_CLOUD_TAGS,
		),
	}
	mng.subscriberDG = mng
	return mng
}

// onResourceUpdated implements SubscriberDataGenerator
func (c *ChPodNSCloudTags) onResourceUpdated(sourceID int, fieldsUpdate *message.PodNamespaceFieldsUpdate, db *mysql.DB) {
	updateInfo := make(map[string]interface{})

	if fieldsUpdate.CloudTags.IsDifferent() {
		bytes, err := json.Marshal(fieldsUpdate.CloudTags.GetNew())
		if err != nil {
			log.Error(err, db.LogPrefixORGID)
			return
		}
		updateInfo["cloud_tags"] = string(bytes)
	}
	if len(updateInfo) > 0 {
		var chItem mysqlmodel.ChPodNSCloudTags
		db.Where("id = ?", sourceID).First(&chItem)
		if chItem.ID == 0 {
			c.SubscriberComponent.dbOperator.add(
				[]CloudTagsKey{{ID: sourceID}},
				[]mysqlmodel.ChPodNSCloudTags{{ID: sourceID, CloudTags: updateInfo["cloud_tags"].(string)}},
				db,
			)
		} else {
			c.SubscriberComponent.dbOperator.update(chItem, updateInfo, CloudTagsKey{ID: sourceID}, db)
		}
	}
}

// onResourceUpdated implements SubscriberDataGenerator
func (c *ChPodNSCloudTags) sourceToTarget(md *message.Metadata, item *mysqlmodel.PodNamespace) (keys []CloudTagsKey, targets []mysqlmodel.ChPodNSCloudTags) {
	if len(item.CloudTags) == 0 {
		return
	}
	bytes, err := json.Marshal(item.CloudTags)
	if err != nil {
		log.Error(err, logger.NewORGPrefix(md.ORGID))
		return
	}
	return []CloudTagsKey{{ID: item.ID}}, []mysqlmodel.ChPodNSCloudTags{{
		ID:          item.ID,
		CloudTags:   string(bytes),
		TeamID:      md.TeamID,
		DomainID:    md.DomainID,
		SubDomainID: md.SubDomainID,
	}}
}

// softDeletedTargetsUpdated implements SubscriberDataGenerator
func (c *ChPodNSCloudTags) softDeletedTargetsUpdated(targets []mysqlmodel.ChPodNSCloudTags, db *mysql.DB) {

}
