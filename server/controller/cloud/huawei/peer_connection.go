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

package huawei

import (
	"fmt"

	cloudcommon "github.com/khulnasoft/deepflow/server/controller/cloud/common"
	"github.com/khulnasoft/deepflow/server/controller/cloud/model"
	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/libs/logger"
)

func (h *HuaWei) getPeerConnections() ([]model.PeerConnection, error) {
	var pns []model.PeerConnection
	for project, token := range h.projectTokenMap {
		jpns, err := h.getRawData(newRawDataGetContext(
			fmt.Sprintf("https://vpc.%s.%s/v2.0/vpc/peerings", project.name, h.config.Domain), token.token, "peerings", pageQueryMethodMarker,
		))
		if err != nil {
			return nil, err
		}

		for i := range jpns {
			jpn := jpns[i]
			if !cloudcommon.CheckJsonAttributes(jpn, []string{"id", "name", "request_vpc_info", "accept_vpc_info"}) {
				continue
			}
			id := common.IDGenerateUUID(h.orgID, jpn.Get("id").MustString())
			name := jpn.Get("name").MustString()
			localTenant := jpn.Get("request_vpc_info").Get("tenant_id").MustString()
			if localTenant == "" {
				log.Infof("exclude peer_connection: %s, missing local region", name, logger.NewORGPrefix(h.orgID))
				continue
			}
			remoteTenant := jpn.Get("accept_vpc_info").Get("tenant_id").MustString()
			if localTenant == "" {
				log.Infof("exclude peer_connection: %s, missing remote region", name, logger.NewORGPrefix(h.orgID))
				continue
			}
			pns = append(
				pns,
				model.PeerConnection{
					Lcuuid:             id,
					Name:               name,
					Label:              id,
					LocalVPCLcuuid:     common.IDGenerateUUID(h.orgID, jpn.Get("request_vpc_info").Get("vpc_id").MustString()),
					RemoteVPCLcuuid:    common.IDGenerateUUID(h.orgID, jpn.Get("accept_vpc_info").Get("vpc_id").MustString()),
					LocalRegionLcuuid:  h.projectNameToRegionLcuuid(localTenant),
					RemoteRegionLcuuid: h.projectNameToRegionLcuuid(remoteTenant),
				},
			)
		}
	}
	return pns, nil
}
