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

package clickhouse

import (
	"sync"

	"github.com/khulnasoft/deepflow/server/libs/lru"
	"github.com/khulnasoft/deepflow/server/querier/common"
	"github.com/khulnasoft/deepflow/server/querier/config"
)

var (
	prometheusSubqueryCacheOnce sync.Once
	prometheusSubqueryCacheIns  *PrometheusSubqueryCache
)

type PrometheusSubqueryCache struct {
	PrometheusSubqueryCache *lru.Cache[common.EntryKey, common.EntryValue]
	Lock                    sync.Mutex
}

func GetPrometheusSubqueryCache() *PrometheusSubqueryCache {
	prometheusSubqueryCacheOnce.Do(func() {
		prometheusSubqueryCacheIns = &PrometheusSubqueryCache{
			PrometheusSubqueryCache: lru.NewCache[common.EntryKey, common.EntryValue](config.Cfg.MaxPrometheusIdSubqueryLruEntry),
		}
	})
	return prometheusSubqueryCacheIns
}

func (c *PrometheusSubqueryCache) Get(key common.EntryKey) (value common.EntryValue, ok bool) {
	c.Lock.Lock()
	value, ok = c.PrometheusSubqueryCache.Get(key)
	c.Lock.Unlock()
	return
}

func (c *PrometheusSubqueryCache) Add(key common.EntryKey, value common.EntryValue) {
	c.Lock.Lock()
	c.PrometheusSubqueryCache.Add(key, value)
	c.Lock.Unlock()
	return
}
