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

package event

import (
	"fmt"

	ctrlrcommon "github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/controller/db/mysql"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
	"github.com/khulnasoft/deepflow/server/controller/recorder/cache/tool"
	rcommon "github.com/khulnasoft/deepflow/server/controller/recorder/common"
	"github.com/khulnasoft/deepflow/server/controller/recorder/constraint"
	"github.com/khulnasoft/deepflow/server/controller/trisolaris/metadata"
	"github.com/khulnasoft/deepflow/server/libs/eventapi"
)

var (
	DESCMigrateFormat     = "%s migrate from %s to %s."
	DESCStateChangeFormat = "%s state changes from %s to %s."
	DESCRecreateFormat    = "%s recreate from %s to %s."
	DESCAddIPFormat       = "%s add ip %s(mac: %s) in subnet %s."
	DESCRemoveIPFormat    = "%s remove ip %s(mac: %s) in subnet %s."
)

type IPTool struct {
	metadata *rcommon.Metadata

	t *tool.DataSet
}

func newTool(t *tool.DataSet) *IPTool {
	return &IPTool{
		metadata: t.GetMetadata(),
		t:        t,
	}
}

func (i *IPTool) GetDeviceOptionsByDeviceID(deviceType, deviceID int) ([]eventapi.TagFieldOption, error) {
	switch deviceType {
	case ctrlrcommon.VIF_DEVICE_TYPE_HOST:
		return i.getHostOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_VM:
		return i.getVMOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_VROUTER:
		return i.getVRouterOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_DHCP_PORT:
		return i.getDHCPPortOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_NAT_GATEWAY:
		return i.getNatGateWayOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_LB:
		return i.getLBOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_RDS_INSTANCE:
		return i.getRDSInstanceOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_REDIS_INSTANCE:
		return i.getRedisInstanceOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_POD_NODE:
		return i.getPodNodeOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_POD_SERVICE:
		return i.getPodServiceOptionsByID(deviceID)
	case ctrlrcommon.VIF_DEVICE_TYPE_POD:
		return i.getPodOptionsByID(deviceID)
	default:
		log.Errorf("device type %d not supported", deviceType, i.metadata.LogPrefixes)
		return nil, fmt.Errorf("device type %d not supported", deviceType)
	}
}

func (i *IPTool) getHostOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetHostInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
	}...)
	return opts, nil
}

func (i *IPTool) getVMOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetVMInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagHostID(info.HostID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_VM),
		eventapi.TagL3DeviceID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getVRouterOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetVRouterInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_VROUTER),
		eventapi.TagL3DeviceID(id),
	}...)

	hostID, ok := i.t.GetHostIDByIP(info.GWLaunchServer)
	if !ok {
		log.Error(idByIPNotFound(ctrlrcommon.RESOURCE_TYPE_HOST_EN, info.GWLaunchServer))
	} else {
		opts = append(opts, []eventapi.TagFieldOption{
			eventapi.TagHostID(hostID),
		}...)
	}

	return opts, nil
}

func (i *IPTool) getDHCPPortOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetDHCPPortInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_DHCP_PORT),
		eventapi.TagL3DeviceID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getNatGateWayOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetNATGatewayInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_NAT_GATEWAY),
		eventapi.TagL3DeviceID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getLBOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetLBInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_LB),
		eventapi.TagL3DeviceID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getRDSInstanceOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetRDSInstanceInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_RDS_INSTANCE),
		eventapi.TagL3DeviceID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getRedisInstanceOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetRedisInstanceInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_REDIS_INSTANCE),
		eventapi.TagL3DeviceID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getPodNodeOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetPodNodeInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagPodClusterID(info.PodClusterID),
		eventapi.TagPodNodeID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getPodServiceOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetPodServiceInfoByID(id)
	if err != nil {
		return nil, err
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_POD_SERVICE),
		eventapi.TagL3DeviceID(id),
		eventapi.TagPodClusterID(info.PodClusterID),
		eventapi.TagPodNSID(info.PodNamespaceID),
		eventapi.TagPodServiceID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getPodOptionsByID(id int) ([]eventapi.TagFieldOption, error) {
	info, err := i.t.GetPodInfoByID(id)
	if err != nil {
		return nil, err
	}
	podGroupType, ok := i.t.GetPodGroupTypeByID(info.PodGroupID)
	if !ok {
		log.Errorf("db pod_group type(id: %d) not found", info.PodGroupID, i.metadata.LogPrefixes)
	}

	var opts []eventapi.TagFieldOption
	opts = append(opts, []eventapi.TagFieldOption{
		eventapi.TagRegionID(info.RegionID),
		eventapi.TagAZID(info.AZID),
		eventapi.TagVPCID(info.VPCID),
		eventapi.TagPodClusterID(info.PodClusterID),
		eventapi.TagPodNSID(info.PodNamespaceID),
		eventapi.TagPodGroupID(info.PodGroupID),
		eventapi.TagPodGroupType(metadata.PodGroupTypeMap[podGroupType]),
		eventapi.TagPodNodeID(info.PodNodeID),
		eventapi.TagPodID(id),
	}...)
	return opts, nil
}

func (i *IPTool) getL3DeviceOptionsByPodNodeID(id int) (opts []eventapi.TagFieldOption, ok bool) {
	vmID, ok := i.t.GetVMIDByPodNodeID(id)
	if ok {
		opts = append(opts, []eventapi.TagFieldOption{eventapi.TagL3DeviceType(ctrlrcommon.VIF_DEVICE_TYPE_VM), eventapi.TagL3DeviceID(vmID)}...)
		vmInfo, err := i.t.GetVMInfoByID(vmID)
		if err != nil {
			log.Error(err)
		} else {
			opts = append(opts, eventapi.TagHostID(vmInfo.HostID))
		}
	}
	return
}

func (i *IPTool) getDeviceNameFromAllByID(deviceType, deviceID int) string {
	switch deviceType {
	case ctrlrcommon.VIF_DEVICE_TYPE_HOST:
		device := findFromAllByID[mysqlmodel.Host](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_HOST_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_VM:
		device := findFromAllByID[mysqlmodel.VM](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_VM_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_VROUTER:
		device := findFromAllByID[mysqlmodel.VRouter](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_VROUTER_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_DHCP_PORT:
		device := findFromAllByID[mysqlmodel.DHCPPort](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_DHCP_PORT_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_NAT_GATEWAY:
		device := findFromAllByID[mysqlmodel.NATGateway](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_NAT_GATEWAY_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_LB:
		device := findFromAllByID[mysqlmodel.LB](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_LB_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_RDS_INSTANCE:
		device := findFromAllByID[mysqlmodel.RDSInstance](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_RDS_INSTANCE_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_REDIS_INSTANCE:
		device := findFromAllByID[mysqlmodel.RedisInstance](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_REDIS_INSTANCE_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_POD_NODE:
		device := findFromAllByID[mysqlmodel.PodNode](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_POD_NODE_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_POD_SERVICE:
		device := findFromAllByID[mysqlmodel.PodService](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_POD_SERVICE_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	case ctrlrcommon.VIF_DEVICE_TYPE_POD:
		device := findFromAllByID[mysqlmodel.Pod](i.metadata.DB, deviceID)
		if device == nil {
			log.Error(dbSoftDeletedResourceByIDNotFound(ctrlrcommon.RESOURCE_TYPE_POD_EN, deviceID), i.metadata.LogPrefixes)
		} else {
			return device.Name
		}
	default:
		log.Errorf("device type: %d is not supported", deviceType, i.metadata.LogPrefixes)
		return ""
	}
	return ""
}

func findFromAllByID[MT constraint.MySQLSoftDeleteModel](db *mysql.DB, id int) *MT {
	var item *MT
	res := db.Unscoped().Where("id = ?", id).Find(&item)
	if res.Error != nil {
		log.Error(dbQueryFailed(res.Error))
		return nil
	}
	if res.RowsAffected != 1 {
		return nil
	}
	return item
}
