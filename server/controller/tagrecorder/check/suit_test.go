/*
 * Copyright (c) 2023 KhulnaSoft, Ltd
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

package tagrecorder

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/khulnasoft/deepflow/server/controller/db/mysql"
	mysqlmodel "github.com/khulnasoft/deepflow/server/controller/db/mysql/model"
)

const (
	TEST_DB_FILE = "./tagrecorder_test.db"
)

type SuiteTest struct {
	suite.Suite
	db *gorm.DB
}

func TestSuite(t *testing.T) {
	if _, err := os.Stat(TEST_DB_FILE); err == nil {
		os.Remove(TEST_DB_FILE)
	}
	mysql.DefaultDB = GetDB(TEST_DB_FILE)
	suite.Run(t, new(SuiteTest))
}

func (t *SuiteTest) SetupSuite() {
	t.db = mysql.DefaultDB

	for _, val := range getModels() {
		t.db.AutoMigrate(val)
	}
}

func (t *SuiteTest) TearDownSuite() {
	sqlDB, _ := t.db.DB()
	sqlDB.Close()

	os.Remove(TEST_DB_FILE)
}

func GetDB(dbFile string) *gorm.DB {
	db, err := gorm.Open(
		sqlite.Open(dbFile),
		&gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}},
	)
	if err != nil {
		fmt.Printf("create sqlite database failed: %s\n", err.Error())
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}

func getModels() []interface{} {
	return []interface{}{
		&mysqlmodel.Region{}, &mysqlmodel.AZ{}, &mysqlmodel.VPC{}, &mysqlmodel.VM{}, &mysqlmodel.VInterface{},
		&mysqlmodel.WANIP{}, &mysqlmodel.LANIP{}, &mysqlmodel.NATGateway{}, &mysqlmodel.NATRule{},
		&mysqlmodel.NATVMConnection{}, &mysqlmodel.LB{}, &mysqlmodel.LBListener{}, &mysqlmodel.LBTargetServer{},
		&mysqlmodel.LBVMConnection{}, &mysqlmodel.PodIngress{}, &mysqlmodel.PodService{}, mysqlmodel.PodGroup{},
		&mysqlmodel.PodGroupPort{}, &mysqlmodel.Pod{},
		&mysqlmodel.ChRegion{}, &mysqlmodel.ChAZ{}, &mysqlmodel.ChVPC{}, &mysqlmodel.ChIPRelation{},
	}
}
