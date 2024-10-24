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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/khulnasoft/deepflow/server/controller/cloud/model"
	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/libs/logger"
)

func (a *Aws) getVPCs(client *ec2.Client) ([]model.VPC, error) {
	log.Debug("get vpcs starting", logger.NewORGPrefix(a.orgID))
	var vpcs []model.VPC

	var retVPCs []types.Vpc
	var nextToken string
	var maxResults int32 = 100
	for {
		var input *ec2.DescribeVpcsInput
		if nextToken == "" {
			input = &ec2.DescribeVpcsInput{MaxResults: &maxResults}
		} else {
			input = &ec2.DescribeVpcsInput{MaxResults: &maxResults, NextToken: &nextToken}
		}
		result, err := client.DescribeVpcs(context.TODO(), input)
		if err != nil {
			log.Errorf("vpc request aws api error: (%s)", err.Error(), logger.NewORGPrefix(a.orgID))
			return []model.VPC{}, err
		}
		retVPCs = append(retVPCs, result.Vpcs...)
		if result.NextToken == nil {
			break
		}
		nextToken = *result.NextToken
	}

	for _, vData := range retVPCs {
		vpcID := a.getStringPointerValue(vData.VpcId)
		vpcName := a.getResultTagName(vData.Tags)
		if vpcName == "" {
			vpcName = vpcID
		}
		vpcLcuuid := common.GetUUIDByOrgID(a.orgID, vpcID)
		vpcs = append(vpcs, model.VPC{
			Lcuuid:       vpcLcuuid,
			Name:         vpcName,
			CIDR:         a.getStringPointerValue(vData.CidrBlock),
			Label:        vpcID,
			RegionLcuuid: a.regionLcuuid,
		})
		a.vpcIDToLcuuid[vpcID] = vpcLcuuid
	}
	log.Debug("get vpcs complete", logger.NewORGPrefix(a.orgID))
	return vpcs, nil
}
