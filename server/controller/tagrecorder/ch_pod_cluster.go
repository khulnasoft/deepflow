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
	"gorm.io/gorm/clause"

	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/controller/db/mysql"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
	"github.com/khulnasoft/deepflow/server/controller/recorder/pubsub/message"
)

type ChPodCluster struct {
	SubscriberComponent[*message.PodClusterFieldsUpdate, message.PodClusterFieldsUpdate, mysqlmodel.PodCluster, mysqlmodel.ChPodCluster, IDKey]
	resourceTypeToIconID map[IconKey]int
}

func NewChPodCluster(resourceTypeToIconID map[IconKey]int) *ChPodCluster {
	mng := &ChPodCluster{
		newSubscriberComponent[*message.PodClusterFieldsUpdate, message.PodClusterFieldsUpdate, mysqlmodel.PodCluster, mysqlmodel.ChPodCluster, IDKey](
			common.RESOURCE_TYPE_POD_CLUSTER_EN, RESOURCE_TYPE_CH_POD_CLUSTER,
		),
		resourceTypeToIconID,
	}
	mng.subscriberDG = mng
	return mng
}

// sourceToTarget implements SubscriberDataGenerator
func (c *ChPodCluster) sourceToTarget(md *message.Metadata, source *mysqlmodel.PodCluster) (keys []IDKey, targets []mysqlmodel.ChPodCluster) {
	iconID := c.resourceTypeToIconID[IconKey{
		NodeType: RESOURCE_TYPE_POD_CLUSTER,
	}]
	sourceName := source.Name
	if source.DeletedAt.Valid {
		sourceName += " (deleted)"
	}

	keys = append(keys, IDKey{ID: source.ID})
	targets = append(targets, mysqlmodel.ChPodCluster{
		ID:          source.ID,
		Name:        sourceName,
		IconID:      iconID,
		TeamID:      md.TeamID,
		DomainID:    md.DomainID,
		SubDomainID: md.SubDomainID,
	})
	return
}

// onResourceUpdated implements SubscriberDataGenerator
func (c *ChPodCluster) onResourceUpdated(sourceID int, fieldsUpdate *message.PodClusterFieldsUpdate, db *mysql.DB) {
	updateInfo := make(map[string]interface{})

	if fieldsUpdate.Name.IsDifferent() {
		updateInfo["name"] = fieldsUpdate.Name.GetNew()
	}
	if len(updateInfo) > 0 {
		var chItem mysqlmodel.ChPodCluster
		db.Where("id = ?", sourceID).First(&chItem)
		c.SubscriberComponent.dbOperator.update(chItem, updateInfo, IDKey{ID: sourceID}, db)
	}
}

// softDeletedTargetsUpdated implements SubscriberDataGenerator
func (c *ChPodCluster) softDeletedTargetsUpdated(targets []mysqlmodel.ChPodCluster, db *mysql.DB) {
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&targets)
}
