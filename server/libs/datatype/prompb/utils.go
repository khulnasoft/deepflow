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

package prompb

import (
	"github.com/khulnasoft/deepflow/server/libs/utils"
)

func (m *WriteRequest) ResetWithBufferReserved() {
	ts := m.Timeseries
	for i := range ts {
		ts[i].Reset()
	}
	ls := m.labelBuffer
	for i := range ls {
		ls[i].Reset()
	}

	m.Reset()

	m.Timeseries = ts[:0]
	m.labelBuffer = ls[:0]
}

func unsafeBytesToString(bytes []byte) string {
	return utils.String(bytes)
}
