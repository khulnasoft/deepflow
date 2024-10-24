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

package genesis

import (
	"time"

	"github.com/khulnasoft/deepflow/server/controller/cloud/model"
	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/libs/logger"
)

func (g *Genesis) getVMs() ([]model.VM, error) {
	log.Debug("get vms starting", logger.NewORGPrefix(g.orgID))
	vms := []model.VM{}
	vmsData := g.genesisData.VMs

	g.cloudStatsd.RefreshAPIMoniter("vms", len(vmsData), time.Time{})

	for _, v := range vmsData {
		launchServer := v.LaunchServer
		if launchServer == "127.0.0.1" {
			launchServer = ""
		}
		vm := model.VM{
			Lcuuid:       v.Lcuuid,
			Name:         v.Name,
			Label:        v.Label,
			HType:        common.VM_HTYPE_VM_C,
			VPCLcuuid:    v.VPCLcuuid,
			State:        int(v.State),
			LaunchServer: launchServer,
			CreatedAt:    v.CreatedAt,
			AZLcuuid:     g.azLcuuid,
			RegionLcuuid: g.regionLcuuid,
		}
		vms = append(vms, vm)
	}
	log.Debug("get vms complete", logger.NewORGPrefix(g.orgID))
	return vms, nil
}
