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
	"math/rand"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	cloudmodel "github.com/khulnasoft/deepflow/server/controller/cloud/model"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
	"github.com/khulnasoft/deepflow/server/controller/recorder/cache"
	"github.com/khulnasoft/deepflow/server/controller/recorder/cache/diffbase"
)

func newCloudHost() cloudmodel.Host {
	lcuuid := uuid.New().String()
	return cloudmodel.Host{
		Lcuuid:   lcuuid,
		Name:     lcuuid[:8],
		VCPUNum:  rand.Intn(10),
		AZLcuuid: uuid.New().String(),
	}
}

func (t *SuiteTest) getHostMock(mockDB bool) (*cache.Cache, cloudmodel.Host) {
	cloudItem := newCloudHost()
	domainLcuuid := uuid.New().String()

	cache_ := cache.NewCache(domainLcuuid)
	if mockDB {
		t.db.Create(&mysqlmodel.Host{Name: cloudItem.Name, Base: mysqlmodel.Base{Lcuuid: cloudItem.Lcuuid}, Domain: domainLcuuid})
		cache_.DiffBaseDataSet.Hosts[cloudItem.Lcuuid] = &diffbase.Host{DiffBase: diffbase.DiffBase{Lcuuid: cloudItem.Lcuuid}, Name: cloudItem.Name}
	}

	cache_.SetSequence(cache_.GetSequence() + 1)

	return cache_, cloudItem
}

func (t *SuiteTest) TestHandleAddHostSucess() {
	cache, cloudItem := t.getHostMock(false)
	assert.Equal(t.T(), len(cache.DiffBaseDataSet.Hosts), 0)

	updater := NewHost(cache, []cloudmodel.Host{cloudItem})
	updater.HandleAddAndUpdate()

	var addedItem *mysqlmodel.Host
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))
	assert.Equal(t.T(), len(cache.DiffBaseDataSet.Hosts), 1)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysqlmodel.Host{})
}

func (t *SuiteTest) TestHandleUpdateHostSucess() {
	cache, cloudItem := t.getHostMock(true)
	cloudItem.Name = cloudItem.Name + "new"
	cloudItem.VCPUNum = cloudItem.VCPUNum + 1
	cloudItem.AZLcuuid = uuid.New().String()

	updater := NewHost(cache, []cloudmodel.Host{cloudItem})
	updater.HandleAddAndUpdate()

	var updatedItem *mysqlmodel.Host
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&updatedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))
	assert.Equal(t.T(), len(cache.DiffBaseDataSet.Hosts), 1)
	assert.Equal(t.T(), updatedItem.Name, cloudItem.Name)
	assert.Equal(t.T(), updatedItem.VCPUNum, cloudItem.VCPUNum)
	assert.Equal(t.T(), updatedItem.AZ, cloudItem.AZLcuuid)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysqlmodel.Host{})
}

func (t *SuiteTest) TestHandleDeleteHostSucess() {
	cache, cloudItem := t.getHostMock(true)
	assert.Equal(t.T(), len(cache.DiffBaseDataSet.Hosts), 1)

	updater := NewHost(cache, []cloudmodel.Host{cloudItem})
	updater.HandleDelete()

	var addedItem *mysqlmodel.Host
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
	assert.Equal(t.T(), len(cache.DiffBaseDataSet.Hosts), 0)
}
