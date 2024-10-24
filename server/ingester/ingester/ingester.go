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

package ingester

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/khulnasoft/deepflow/server/ingester/app_log"
	"github.com/khulnasoft/deepflow/server/ingester/ckmonitor"
	"github.com/khulnasoft/deepflow/server/ingester/datasource"
	"github.com/khulnasoft/deepflow/server/ingester/exporters"
	"github.com/khulnasoft/deepflow/server/libs/grpc"
	"github.com/khulnasoft/deepflow/server/libs/logger"
	"github.com/khulnasoft/deepflow/server/libs/pool"
	"github.com/khulnasoft/deepflow/server/libs/receiver"
	"github.com/khulnasoft/deepflow/server/libs/stats"

	logging "github.com/op/go-logging"
	yaml "gopkg.in/yaml.v2"

	servercommon "github.com/khulnasoft/deepflow/server/common"
	applicationlogcfg "github.com/khulnasoft/deepflow/server/ingester/app_log/config"
	"github.com/khulnasoft/deepflow/server/ingester/ckissu"
	"github.com/khulnasoft/deepflow/server/ingester/common"
	"github.com/khulnasoft/deepflow/server/ingester/config"
	eventcfg "github.com/khulnasoft/deepflow/server/ingester/event/config"
	"github.com/khulnasoft/deepflow/server/ingester/event/event"
	exporterscfg "github.com/khulnasoft/deepflow/server/ingester/exporters/config"
	extmetricscfg "github.com/khulnasoft/deepflow/server/ingester/ext_metrics/config"
	"github.com/khulnasoft/deepflow/server/ingester/ext_metrics/ext_metrics"
	flowlogcfg "github.com/khulnasoft/deepflow/server/ingester/flow_log/config"
	flowlog "github.com/khulnasoft/deepflow/server/ingester/flow_log/flow_log"
	flowmetricscfg "github.com/khulnasoft/deepflow/server/ingester/flow_metrics/config"
	flowmetrics "github.com/khulnasoft/deepflow/server/ingester/flow_metrics/flow_metrics"
	pcapcfg "github.com/khulnasoft/deepflow/server/ingester/pcap/config"
	"github.com/khulnasoft/deepflow/server/ingester/pcap/pcap"
	profilecfg "github.com/khulnasoft/deepflow/server/ingester/profile/config"
	"github.com/khulnasoft/deepflow/server/ingester/profile/profile"
	prometheuscfg "github.com/khulnasoft/deepflow/server/ingester/prometheus/config"
	"github.com/khulnasoft/deepflow/server/ingester/prometheus/prometheus"
)

var log = logging.MustGetLogger("ingester")

const (
	PROFILER_PORT                = 9526
	MAX_SLAVE_PLATFORMDATA_COUNT = 128
)

func Start(configPath string, shared *servercommon.ControllerIngesterShared) []io.Closer {
	cfg := config.Load(configPath)
	bytes, _ := yaml.Marshal(cfg)

	logger.EnableStdoutLog()
	logger.EnableFileLog(cfg.LogFile)
	logLevel, _ := logging.LogLevel(cfg.LogLevel)
	logging.SetLevel(logLevel, "")

	log.Info("==================== Launching DeepFlow-Server-Ingester ====================")
	log.Infof("ingester base config:\n%s", string(bytes))

	pool.SetCounterRegisterCallback(func(counter *pool.Counter) {
		tags := stats.OptionStatTags{
			"name":                counter.Name,
			"object_size":         strconv.Itoa(int(counter.ObjectSize)),
			"pool_size_per_cpu":   strconv.Itoa(int(counter.PoolSizePerCPU)),
			"init_full_pool_size": strconv.Itoa(int(counter.InitFullPoolSize)),
		}
		common.RegisterCountableForIngester("pool", counter, tags)
	})
	stats.SetHostname(cfg.MyNodeName)
	stats.RegisterGcMonitor()
	stats.SetMinInterval(time.Duration(cfg.StatsInterval) * time.Second)
	stats.SetRemoteType(stats.REMOTE_TYPE_DFSTATSD)
	stats.SetDFRemote(net.JoinHostPort("127.0.0.1", strconv.Itoa(int(cfg.ListenPort))))

	receiver := receiver.NewReceiver(int(cfg.ListenPort), cfg.UDPReadBuffer, cfg.TCPReadBuffer, cfg.TCPReaderBuffer)

	ingesterOrgHandler := NewOrgHandler(cfg)
	closers := []io.Closer{}

	if cfg.IngesterEnabled {
		flowLogConfig := flowlogcfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(flowLogConfig)
		log.Infof("flow log config:\n%s", string(bytes))

		flowMetricsConfig := flowmetricscfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(flowMetricsConfig)
		log.Infof("flow metrics config:\n%s", string(bytes))

		extMetricsConfig := extmetricscfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(extMetricsConfig)
		log.Infof("ext_metrics config:\n%s", string(bytes))

		eventConfig := eventcfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(eventConfig)
		log.Infof("event config:\n%s", string(bytes))

		pcapConfig := pcapcfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(pcapConfig)
		log.Infof("pcap config:\n%s", string(bytes))

		profileConfig := profilecfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(profileConfig)
		log.Infof("profile config:\n%s", string(bytes))

		prometheusConfig := prometheuscfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(prometheusConfig)
		log.Infof("prometheus config:\n%s", string(bytes))

		applicationLogConfig := applicationlogcfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(applicationLogConfig)
		log.Infof("application log  config:\n%s", string(bytes))

		exportersConfig := exporterscfg.Load(cfg, configPath)
		bytes, _ = yaml.Marshal(exportersConfig)
		log.Infof("exporters config:\n%s", string(bytes))

		var issu *ckissu.Issu
		if !cfg.StorageDisabled {
			var err error
			// 创建、修改、删除数据源及其存储时长
			ds := datasource.NewDatasourceManager(cfg, flowMetricsConfig.CKReadTimeout)
			ds.Start()
			closers = append(closers, ds)

			// clickhouse表结构变更处理
			issu, err = ckissu.NewCKIssu(cfg)
			checkError(err)
			// If there is a table name change, do the table name update first
			err = issu.RunRenameTable(ds)
			checkError(err)
		}

		// platformData manager init
		controllers := make([]net.IP, len(cfg.ControllerIPs))
		for i, ipString := range cfg.ControllerIPs {
			controllers[i] = net.ParseIP(ipString)
			if controllers[i].To4() != nil {
				controllers[i] = controllers[i].To4()
			}
		}
		platformDataManager := grpc.NewPlatformDataManager(
			controllers,
			int(cfg.ControllerPort),
			MAX_SLAVE_PLATFORMDATA_COUNT,
			cfg.GrpcBufferSize,
			cfg.NodeIP,
			receiver)

		exporters := exporters.NewExporters(exportersConfig)
		if exporters != nil {
			exporters.Start()
			closers = append(closers, exporters)
		}

		// 写流日志数据
		flowLog, err := flowlog.NewFlowLog(flowLogConfig, shared.TraceTreeQueue, receiver, platformDataManager, exporters)
		checkError(err)
		flowLog.Start()
		closers = append(closers, flowLog)

		if !cfg.StorageDisabled {
			// 写ext_metrics数据
			extMetrics, err := ext_metrics.NewExtMetrics(extMetricsConfig, receiver, platformDataManager)
			checkError(err)
			extMetrics.Start()
			closers = append(closers, extMetrics)

			// 写遥测数据
			flowMetrics, err := flowmetrics.NewFlowMetrics(flowMetricsConfig, receiver, platformDataManager, exporters)
			checkError(err)
			flowMetrics.Start()
			closers = append(closers, flowMetrics)

			// write event data
			event, err := event.NewEvent(eventConfig, shared.ResourceEventQueue, receiver, platformDataManager, exporters)
			checkError(err)
			event.Start()
			closers = append(closers, event)

			// write pcap data
			pcaper, err := pcap.NewPcaper(receiver, pcapConfig)
			checkError(err)
			pcaper.Start()
			closers = append(closers, pcaper)

			// write profile data
			profile, err := profile.NewProfile(profileConfig, receiver, platformDataManager)
			checkError(err)
			profile.Start()
			closers = append(closers, profile)

			// write prometheus data
			prometheus, err := prometheus.NewPrometheusHandler(prometheusConfig, receiver, platformDataManager)
			checkError(err)
			prometheus.Start()
			closers = append(closers, prometheus)
			ingesterOrgHandler.SetPromHandler(prometheus)

			// write application log data
			applicationLog, err := app_log.NewApplicationLogger(applicationLogConfig, receiver, platformDataManager)
			checkError(err)
			applicationLog.Start()
			closers = append(closers, applicationLog)

			// 检查clickhouse的磁盘空间占用，达到阈值时，自动删除老数据
			cm, err := ckmonitor.NewCKMonitor(cfg)
			checkError(err)
			cm.Start()
			closers = append(closers, cm)

			// 初始化建表完成,再执行issu
			time.Sleep(time.Second)
			err = issu.Start()
			checkError(err)
			// after issu execution is completed, should close it to prevent the connection from occupying memory.
			issu.Close()
			issu = nil
		}
	}
	// receiver后启动，防止启动后收到数据无法处理，而上报异常日志
	receiver.Start()
	closers = append(closers, receiver)
	servercommon.SetOrgHandler(ingesterOrgHandler)

	return closers
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
