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

package aws

import (
	"github.com/khulnasoft/deepflow/server/controller/cloud/model"
	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/libs/logger"
)

func (a *Aws) getFloatingIPs() (floatingIPs []model.FloatingIP, err error) {
	log.Debug("get floating ips starting", logger.NewORGPrefix(a.orgID))
	for ip, v := range a.publicIPToVinterface {
		floatingIP := model.FloatingIP{
			Lcuuid:        common.GetUUIDByOrgID(a.orgID, v.Lcuuid+ip),
			IP:            ip,
			VMLcuuid:      v.DeviceLcuuid,
			NetworkLcuuid: common.NETWORK_ISP_LCUUID,
			VPCLcuuid:     v.VPCLcuuid,
			RegionLcuuid:  v.RegionLcuuid,
		}
		floatingIPs = append(floatingIPs, floatingIP)
	}
	log.Debug("get floating ips complete", logger.NewORGPrefix(a.orgID))
	return
}
