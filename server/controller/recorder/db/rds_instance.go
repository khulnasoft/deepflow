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

package db

import (
	ctrlrcommon "github.com/khulnasoft/deepflow/server/controller/common"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
)

type RDSInstance struct {
	OperatorBase[*mysqlmodel.RDSInstance, mysqlmodel.RDSInstance]
}

func NewRDSInstance() *RDSInstance {
	operater := &RDSInstance{
		newOperatorBase[*mysqlmodel.RDSInstance](
			ctrlrcommon.RESOURCE_TYPE_RDS_INSTANCE_EN,
			true,
			true,
		),
	}
	return operater
}
