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

package listener

import (
	cloudmodel "github.com/deepflowio/deepflow/server/controller/cloud/model"
	mysqlmodel "github.com/deepflowio/deepflow/server/controller/db/mysql/model"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache/diffbase"
)

type Subnet struct {
	cache *cache.Cache
}

func NewSubnet(c *cache.Cache) *Subnet {
	return &Subnet{
		cache: c,
	}
}

func (s *Subnet) OnUpdaterAdded(addedDBItems []*mysqlmodel.Subnet) {
	s.cache.AddSubnets(addedDBItems)
}

func (s *Subnet) OnUpdaterUpdated(cloudItem *cloudmodel.Subnet, diffBase *diffbase.Subnet) {
	diffBase.Update(cloudItem)
}

func (s *Subnet) OnUpdaterDeleted(lcuuids []string) {
	s.cache.DeleteSubnets(lcuuids)
}
