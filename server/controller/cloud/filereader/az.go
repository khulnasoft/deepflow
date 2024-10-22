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

package filereader

import (
	"github.com/khulnasoft/deepflow/server/controller/cloud/model"
	"github.com/khulnasoft/deepflow/server/controller/common"
)

func (f *FileReader) getAZs(fileInfo *FileInfo) ([]model.AZ, error) {
	var retAZs []model.AZ

	for _, az := range fileInfo.AZs {
		regionLcuuid, err := f.getRegionLcuuid(az.Region)
		if err != nil {
			return nil, err
		}

		lcuuid := common.GenerateUUIDByOrgID(f.orgID, f.UuidGenerate+"_az_"+az.Name)
		f.azNameToLcuuid[az.Name] = lcuuid
		retAZs = append(retAZs, model.AZ{
			Lcuuid:       lcuuid,
			Name:         az.Name,
			RegionLcuuid: regionLcuuid,
		})
	}
	return retAZs, nil
}
