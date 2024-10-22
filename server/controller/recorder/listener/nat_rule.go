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

package listener

import (
	cloudmodel "github.com/khulnasoft/deepflow/server/controller/cloud/model"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
	"github.com/khulnasoft/deepflow/server/controller/recorder/cache"
	"github.com/khulnasoft/deepflow/server/controller/recorder/cache/diffbase"
)

type NATRule struct {
	cache *cache.Cache
}

func NewNATRule(c *cache.Cache) *NATRule {
	listener := &NATRule{
		cache: c,
	}
	return listener
}

func (r *NATRule) OnUpdaterAdded(addedDBItems []*mysqlmodel.NATRule) {
	r.cache.AddNATRules(addedDBItems)
}

func (r *NATRule) OnUpdaterUpdated(cloudItem *cloudmodel.NATRule, diffBase *diffbase.NATRule) {
}

func (r *NATRule) OnUpdaterDeleted(lcuuids []string) {
	r.cache.DeleteNATRules(lcuuids)
}
