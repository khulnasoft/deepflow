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

package controller

import (
	"context"
	"flag"
	"os"
	"strconv"
	"time"

	logging "github.com/op/go-logging"
	yaml "gopkg.in/yaml.v2"

	servercommon "github.com/khulnasoft/deepflow/server/common"
	"github.com/khulnasoft/deepflow/server/controller/common"
	"github.com/khulnasoft/deepflow/server/controller/config"
	"github.com/khulnasoft/deepflow/server/controller/db/mysql"
	"github.com/khulnasoft/deepflow/server/controller/db/redis"
	"github.com/khulnasoft/deepflow/server/controller/election"
	"github.com/khulnasoft/deepflow/server/controller/genesis"
	"github.com/khulnasoft/deepflow/server/controller/grpc"
	_ "github.com/khulnasoft/deepflow/server/controller/grpc/controller"
	_ "github.com/khulnasoft/deepflow/server/controller/grpc/synchronizer"
	"github.com/khulnasoft/deepflow/server/controller/http"
	"github.com/khulnasoft/deepflow/server/controller/http/router"
	"github.com/khulnasoft/deepflow/server/controller/manager"
	"github.com/khulnasoft/deepflow/server/controller/monitor"
	"github.com/khulnasoft/deepflow/server/controller/prometheus"
	"github.com/khulnasoft/deepflow/server/controller/recorder"
	"github.com/khulnasoft/deepflow/server/controller/report"
	"github.com/khulnasoft/deepflow/server/controller/statsd"
	"github.com/khulnasoft/deepflow/server/controller/tagrecorder"
	"github.com/khulnasoft/deepflow/server/controller/trisolaris"
	_ "github.com/khulnasoft/deepflow/server/controller/trisolaris/services/grpc/debug"
	_ "github.com/khulnasoft/deepflow/server/controller/trisolaris/services/grpc/healthcheck"
	_ "github.com/khulnasoft/deepflow/server/controller/trisolaris/services/http/cache"
	_ "github.com/khulnasoft/deepflow/server/controller/trisolaris/services/http/upgrade"
)

var log = logging.MustGetLogger("controller")

type Controller struct{}

func Start(ctx context.Context, configPath, serverLogFile string, shared *servercommon.ControllerIngesterShared) {
	common.InitEnvData()
	flag.Parse()

	serverCfg := config.DefaultConfig()
	serverCfg.Load(configPath)
	cfg := &serverCfg.ControllerConfig
	bytes, _ := yaml.Marshal(cfg)
	log.Info("==================== Launching DeepFlow-Server-Controller ====================")
	log.Infof("controller config:\n%s", string(bytes))
	setGlobalConfig(cfg)

	httpServer := http.NewServer(serverLogFile, cfg)
	httpServer.Start()

	defer router.SetInitStageForHealthChecker(router.OK)

	router.SetInitStageForHealthChecker("Election init")
	// start election
	if common.IsStandaloneRunningMode() == false {
		// in standalone mode, We have no way to elect because there is no k8s module
		go election.Start(ctx, cfg)
	}

	isMasterController := IsMasterController(cfg)
	if isMasterController {
		router.SetInitStageForHealthChecker("MySQL migration")
		migrateMySQL(cfg)
	}

	router.SetInitStageForHealthChecker("MySQL init")
	// 初始化MySQL
	if err := mysql.GetDBs().Init(cfg.MySqlCfg); err != nil {
		log.Errorf("init mysql failed: %s", err.Error())
		time.Sleep(time.Second)
		os.Exit(0)
	}

	// 启动资源ID管理器
	router.SetInitStageForHealthChecker("Resource ID manager init")
	recorderResource := recorder.GetResource().Init(ctx, cfg.ManagerCfg.TaskCfg.RecorderCfg)
	if isMasterController {
		err := recorderResource.IDManagers.Start(ctx)
		if err != nil {
			log.Errorf("resource id manager start failed: %s", err.Error())
			time.Sleep(time.Second)
			os.Exit(0)
		}
	}

	// 初始化Redis
	if cfg.RedisCfg.Enabled {
		router.SetInitStageForHealthChecker("Redis init")

		err := redis.Init(ctx, cfg.RedisCfg)
		if err != nil {
			log.Errorf("connect redis failed: %s", err.Error())
			time.Sleep(time.Second)
			os.Exit(0)
		}
	}

	router.SetInitStageForHealthChecker("Statsd init")
	// start statsd
	statsd.NewStatsdMonitor(cfg.StatsdCfg)

	router.SetInitStageForHealthChecker("Genesis init")
	// 启动genesis
	g := genesis.NewGenesis(ctx, cfg)

	// start tagrecorder before manager to prevent recorder from publishing message when tagrecorder is not ready
	router.SetInitStageForHealthChecker("TagRecorder init")
	tr := tagrecorder.GetSingleton()
	tr.Init(ctx, *cfg)
	if err := tr.SubscriberManager.Start(); err != nil {
		log.Errorf("get icon failed: %s", err.Error())
		time.Sleep(time.Second)
		os.Exit(0)
	}

	router.SetInitStageForHealthChecker("Manager init")
	// 启动resource manager
	// 每个云平台启动一个cloud和recorder
	m := manager.NewManager(cfg.ManagerCfg, shared.ResourceEventQueue)
	m.Start()

	router.SetInitStageForHealthChecker("Trisolaris init")
	// 启动trisolaris
	tm := trisolaris.NewTrisolarisManager(&cfg.TrisolarisCfg, mysql.DefaultDB.DB)
	go tm.Start()

	router.SetInitStageForHealthChecker("Prometheus init")
	prometheus := prometheus.GetSingleton()
	prometheus.SynchronizerCaches.Start(ctx, &cfg.PrometheusCfg)
	prometheus.Encoders.Init(ctx, cfg.PrometheusCfg)
	prometheus.Clear.Init(ctx, cfg.PrometheusCfg)
	// prometheus.APPLabelLayoutUpdater.Init(ctx, &cfg.PrometheusCfg)
	if isMasterController {
		prometheus.Encoders.Start(ctx)
	}

	go checkAndStartAllRegionMasterFunctions(ctx)

	router.SetInitStageForHealthChecker("Master function init")
	controllerCheck := monitor.NewControllerCheck(cfg, ctx)
	analyzerCheck := monitor.NewAnalyzerCheck(cfg, ctx)
	go checkAndStartMasterFunctions(cfg, ctx, controllerCheck, analyzerCheck)

	router.SetInitStageForHealthChecker("Register routers init")
	httpServer.SetControllerChecker(controllerCheck)
	httpServer.SetAnalyzerChecker(analyzerCheck)
	httpServer.SetGenesis(g)
	httpServer.SetManager(m)
	httpServer.RegisterRouters()

	grpcStart(ctx, cfg)

	if !cfg.ReportingDisabled {
		go report.NewReportServer(mysql.DefaultDB.DB).StartReporting()
	}
}

func grpcStart(ctx context.Context, cfg *config.ControllerConfig) {
	go grpc.Run(ctx, cfg)
	_, err1 := os.Stat(cfg.AgentSSLKeyFile)
	_, err2 := os.Stat(cfg.AgentSSLCertFile)
	if err1 == nil && err2 == nil {
		go grpc.RunTLS(ctx, cfg)
	}
}

func setGlobalConfig(cfg *config.ControllerConfig) {
	grpcPort, err := strconv.Atoi(cfg.GrpcPort)
	if err != nil {
		log.Error("config grpc-port is not a port")
		time.Sleep(time.Second)
		os.Exit(0)
	}
	grpcNodePort, err := strconv.Atoi(cfg.GrpcNodePort)
	if err != nil {
		log.Error("config grpc-node-port is not a port")
		time.Sleep(time.Second)
		os.Exit(0)
	}
	common.GConfig = &common.GlobalConfig{
		HTTPPort:     cfg.ListenPort,
		HTTPNodePort: cfg.ListenNodePort,
		GRPCPort:     grpcPort,
		GRPCNodePort: grpcNodePort,
	}
}
