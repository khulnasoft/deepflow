/**
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

package tagrecorder

import (
	"sync"

	"golang.org/x/exp/slices"

	"github.com/khulnasoft/deepflow/server/controller/config"
	"github.com/khulnasoft/deepflow/server/controller/db/mysql"
	"github.com/khulnasoft/deepflow/server/controller/recorder/constraint"
	"github.com/khulnasoft/deepflow/server/controller/recorder/pubsub"
	"github.com/khulnasoft/deepflow/server/controller/recorder/pubsub/message"
	msgconstraint "github.com/khulnasoft/deepflow/server/controller/recorder/pubsub/message/constraint"
	"github.com/khulnasoft/deepflow/server/libs/logger"
)

var (
	subscriberManagerOnce sync.Once
	subscriberManager     *SubscriberManager
)

type SubscriberManager struct {
	cfg                  config.ControllerConfig
	domainLcuuidToIconID map[string]int
	resourceTypeToIconID map[IconKey]int

	subscribers []Subscriber
}

func GetSubscriberManager() *SubscriberManager {
	subscriberManagerOnce.Do(func() {
		subscriberManager = &SubscriberManager{}
	})
	return subscriberManager
}

func (c *SubscriberManager) Init(cfg config.ControllerConfig) {
	c.cfg = cfg
}

func (c *SubscriberManager) Start() (err error) {
	log.Info("tagrecorder subscriber manager started")
	c.domainLcuuidToIconID, c.resourceTypeToIconID, err = GetIconInfo(c.cfg)
	if err != nil {
		return err
	}
	c.subscribers = c.getSubscribers()
	log.Infof("tagrecorder run start")
	for _, subscriber := range c.subscribers {
		subscriber.SetConfig(c.cfg)
		subscriber.Subscribe()
	}
	return nil
}

func (m *SubscriberManager) GetSubscribers(subResourceType string) []Subscriber {
	ss := make([]Subscriber, 0)
	if subResourceType == pubsub.PubSubTypeDomain {
		return m.subscribers
	} else if subResourceType == pubsub.PubSubTypeSubDomain {
		for _, s := range m.subscribers {
			if slices.Contains(SUB_DOMAIN_RESOURCE_TYPES, s.GetSubResourceType()) {
				ss = append(ss, s)
			}
		}
	} else {
		for _, s := range m.subscribers {
			if s.GetSubResourceType() == subResourceType {
				ss = append(ss, s)
			}
		}
	}
	return ss
}

func (c *SubscriberManager) getSubscribers() []Subscriber {
	subscribers := []Subscriber{
		NewChAZ(c.domainLcuuidToIconID, c.resourceTypeToIconID),
		NewChVMDevice(c.resourceTypeToIconID),
		NewChHostDevice(c.resourceTypeToIconID),
		NewChVRouterDevice(c.resourceTypeToIconID),
		NewChDHCPPortDevice(c.resourceTypeToIconID),
		NewChNATGatewayDevice(c.resourceTypeToIconID),
		NewChLBDevice(c.resourceTypeToIconID),
		NewChRDSInstanceDevice(c.resourceTypeToIconID),
		NewChRedisInstanceDevice(c.resourceTypeToIconID),
		NewChPodServiceDevice(c.resourceTypeToIconID),
		NewChPodDevice(c.resourceTypeToIconID),
		NewChPodGroupDevice(c.resourceTypeToIconID),
		NewChPodNodeDevice(c.resourceTypeToIconID),
		NewChPodClusterDevice(c.resourceTypeToIconID),
		NewChProcessDevice(c.resourceTypeToIconID),
		NewChOSAppTag(),
		NewChOSAppTags(),
		NewChPodK8sLabel(),
		NewChPodK8sLabels(),
		NewChPodK8sAnnotation(),
		NewChPodK8sAnnotations(),
		NewChPodK8sEnv(),
		NewChPodK8sEnvs(),
		NewChChostCloudTag(),
		NewChChostCloudTags(),
		NewChNetwork(c.resourceTypeToIconID),
		NewChChost(),
		NewChGProcess(c.resourceTypeToIconID),
		NewChVPC(c.resourceTypeToIconID),
		NewChPodCluster(c.resourceTypeToIconID),
		NewChPod(c.resourceTypeToIconID),
		NewChPodGroup(c.resourceTypeToIconID),
		NewChPodIngress(),
		NewChPodNode(c.resourceTypeToIconID),
		NewChPodNamespace(c.resourceTypeToIconID),
		NewChPodService(),

		NewChPodServiceK8sAnnotation(),
		NewChPodServiceK8sAnnotations(),
		NewChPodNSCloudTag(),
		NewChPodNSCloudTags(),
		NewChPodServiceK8sLabel(),
		NewChPodServiceK8sLabels(),
	}
	return subscribers
}

type Subscriber interface {
	Subscribe()
	SetConfig(config.ControllerConfig)
	GetSubResourceType() string
	pubsub.ResourceBatchAddedSubscriber
	pubsub.ResourceUpdatedSubscriber
	pubsub.ResourceBatchDeletedSubscriber
	OnDomainDeleted(md *message.Metadata)
	OnSubDomainDeleted(md *message.Metadata)
	OnSubDomainTeamIDUpdated(md *message.Metadata)
	ResourceUpdateAtInfoUpdated(md *message.Metadata, db *mysql.DB)
}

type SubscriberDataGenerator[MUPT msgconstraint.FieldsUpdatePtr[MUT], MUT msgconstraint.FieldsUpdate, MT constraint.MySQLModel, CT MySQLChModel, KT ChModelKey] interface {
	sourceToTarget(md *message.Metadata, resourceMySQLItem *MT) (chKeys []KT, chItems []CT) // 将源表数据转换为CH表数据
	onResourceUpdated(int, MUPT, *mysql.DB)
	softDeletedTargetsUpdated([]CT, *mysql.DB)
}

type SubscriberComponent[MUPT msgconstraint.FieldsUpdatePtr[MUT], MUT msgconstraint.FieldsUpdate, MT constraint.MySQLModel, CT MySQLChModel, KT ChModelKey] struct {
	cfg config.ControllerConfig

	subResourceTypeName string // 订阅表资源类型，即源表资源类型
	resourceTypeName    string // CH表资源类型
	dbOperator          operator[CT, KT]
	subscriberDG        SubscriberDataGenerator[MUPT, MUT, MT, CT, KT]
}

func newSubscriberComponent[MUPT msgconstraint.FieldsUpdatePtr[MUT], MUT msgconstraint.FieldsUpdate, MT constraint.MySQLModel, CT MySQLChModel, KT ChModelKey](
	sourceResourceTypeName, resourceTypeName string,
) SubscriberComponent[MUPT, MUT, MT, CT, KT] {
	s := SubscriberComponent[MUPT, MUT, MT, CT, KT]{
		subResourceTypeName: sourceResourceTypeName,
		resourceTypeName:    resourceTypeName,
	}
	s.initDBOperator()
	return s
}

func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) GetSubResourceType() string {
	return s.subResourceTypeName
}

func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) initDBOperator() {
	s.dbOperator = newOperator[CT, KT](s.resourceTypeName)
}

func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) generateKeyTargets(md *message.Metadata, sources []*MT) ([]KT, []CT) {
	keys := []KT{}
	targets := []CT{}
	for _, item := range sources {
		ks, ts := s.subscriberDG.sourceToTarget(md, item)
		keys = append(keys, ks...)
		targets = append(targets, ts...)
	}
	return keys, targets
}

func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) SetConfig(cfg config.ControllerConfig) {
	s.cfg = cfg
	s.dbOperator.setConfig(cfg)
}

func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) SetIconInfo(domainLcuuidToIconID map[string]int, resourceTypeToIconID map[IconKey]int) {

}

func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) Subscribe() {
	pubsub.Subscribe(s.subResourceTypeName, pubsub.TopicResourceBatchAddedMySQL, s)
	pubsub.Subscribe(s.subResourceTypeName, pubsub.TopicResourceUpdatedFields, s)
	pubsub.Subscribe(s.subResourceTypeName, pubsub.TopicResourceBatchDeletedMySQL, s)
}

// OnResourceBatchAdded implements interface Subscriber in recorder/pubsub/subscriber.go
func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) OnResourceBatchAdded(md *message.Metadata, msg interface{}) { // TODO handle org
	items := msg.([]*MT)
	db, err := mysql.GetDB(md.ORGID)
	if err != nil {
		log.Error("get org dbinfo fail", logger.NewORGPrefix(md.ORGID))
	}
	keys, chItems := s.generateKeyTargets(md, items)
	s.dbOperator.batchPage(keys, chItems, s.dbOperator.add, db)
}

// OnResourceBatchUpdated implements interface Subscriber in recorder/pubsub/subscriber.go
func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) OnResourceUpdated(md *message.Metadata, msg interface{}) {
	updateFields := msg.(MUPT)
	db, err := mysql.GetDB(md.ORGID)
	if err != nil {
		log.Error("get org dbinfo fail", logger.NewORGPrefix(md.ORGID))
	}
	s.subscriberDG.onResourceUpdated(updateFields.GetID(), updateFields, db)
}

// OnResourceBatchDeleted implements interface Subscriber in recorder/pubsub/subscriber.go
func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) OnResourceBatchDeleted(md *message.Metadata, msg interface{}) {
	items := msg.([]*MT)
	db, err := mysql.GetDB(md.ORGID)
	if err != nil {
		log.Error("get org dbinfo fail", logger.NewORGPrefix(md.ORGID))
	}
	keys, chItems := s.generateKeyTargets(md, items)
	if md.SoftDelete {
		s.subscriberDG.softDeletedTargetsUpdated(chItems, db)
	} else {
		s.dbOperator.batchPage(keys, chItems, s.dbOperator.delete, db)
		s.ResourceUpdateAtInfoUpdated(md, db)
	}

}

// Delete resource by domain
func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) OnDomainDeleted(md *message.Metadata) {
	var chModel CT
	db, err := mysql.GetDB(md.ORGID)
	if err != nil {
		log.Error("get org dbinfo fail", logger.NewORGPrefix(md.ORGID))
	}
	if err := db.Where("domain_id = ?", md.DomainID).Delete(&chModel).Error; err != nil {
		log.Error(err, logger.NewORGPrefix(md.ORGID))
	}
	s.ResourceUpdateAtInfoUpdated(md, db)
}

// Delete resource by sub domain
func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) OnSubDomainDeleted(md *message.Metadata) {
	var chModel CT
	db, err := mysql.GetDB(md.ORGID)
	if err != nil {
		log.Error("get org dbinfo fail", logger.NewORGPrefix(md.ORGID))
	}
	if err := db.Where("sub_domain_id = ?", md.SubDomainID).Delete(&chModel).Error; err != nil {
		log.Error(err, logger.NewORGPrefix(md.ORGID))
	}
	s.ResourceUpdateAtInfoUpdated(md, db)
}

// Update team_id of resource by sub domain
func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) OnSubDomainTeamIDUpdated(md *message.Metadata) {
	var chModel CT
	db, err := mysql.GetDB(md.ORGID)
	if err != nil {
		log.Error("get org dbinfo fail", logger.NewORGPrefix(md.ORGID))
	}
	if err := db.Model(&chModel).Where("sub_domain_id = ?", md.SubDomainID).Update("team_id", md.TeamID).Error; err != nil {
		log.Error(err, db.LogPrefixORGID)
	}
}

// Update updated_at when resource is deleted
func (s *SubscriberComponent[MUPT, MUT, MT, CT, KT]) ResourceUpdateAtInfoUpdated(md *message.Metadata, db *mysql.DB) {
	var updateItems []MT
	err := db.Unscoped().First(&updateItems).Error
	if err == nil {
		db.Save(updateItems)
	}
}
