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

package baidubce

import (
	"time"

	"github.com/baidubce/bce-sdk-go/services/vpc"

	"github.com/khulnasoft/deepflow/server/controller/cloud/model"
	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/libs/logger"
)

func (b *BaiduBce) getVPCs() ([]model.VPC, map[string]string, map[string]string, error) {
	var retVPCs []model.VPC
	var vpcIdToLcuuid map[string]string
	var vpcIdToName map[string]string

	log.Debug("get vpcs starting", logger.NewORGPrefix(b.orgID))

	vpcClient, _ := vpc.NewClient(b.secretID, b.secretKey, "bcc."+b.endpoint)
	vpcClient.Config.ConnectionTimeoutInMillis = b.httpTimeout * 1000
	marker := ""
	args := &vpc.ListVPCArgs{}
	results := make([]*vpc.ListVPCResult, 0)
	for {
		args.Marker = marker
		startTime := time.Now()
		result, err := vpcClient.ListVPC(args)
		if err != nil {
			log.Error(err, logger.NewORGPrefix(b.orgID))
			return nil, nil, nil, err
		}
		b.cloudStatsd.RefreshAPIMoniter("ListVPC", len(result.VPCs), startTime)
		results = append(results, result)
		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}

	b.debugger.WriteJson("ListVPC", " ", structToJson(results))
	vpcIdToName = make(map[string]string)
	vpcIdToLcuuid = make(map[string]string)
	for _, r := range results {
		for _, vpc := range r.VPCs {
			vpcLcuuid := common.GenerateUUIDByOrgID(b.orgID, vpc.VPCID)
			retVPC := model.VPC{
				Lcuuid:       vpcLcuuid,
				Name:         vpc.Name,
				CIDR:         vpc.Cidr,
				RegionLcuuid: b.regionLcuuid,
			}
			retVPCs = append(retVPCs, retVPC)
			vpcIdToName[vpc.VPCID] = vpc.Name
			vpcIdToLcuuid[vpc.VPCID] = vpcLcuuid
		}
	}
	log.Debug("get vpcs complete", logger.NewORGPrefix(b.orgID))
	return retVPCs, vpcIdToLcuuid, vpcIdToName, nil
}
