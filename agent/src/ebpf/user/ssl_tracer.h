/*
 * Copyright (c) 2022 KhulnaSoft, Ltd
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

#ifndef _BPF_SSL_TRACER_H_
#define _BPF_SSL_TRACER_H_

#include "tracer.h"

// Scan /proc/ to get all processes when the agent starts
int collect_ssl_uprobe_syms_from_procfs(struct tracer_probes_conf *conf);

// Get the process creation event and put the event into the queue
void ssl_process_exec(int pid);

// Process events in the queue
void ssl_events_handle(void);

// Process exit, reclaim resources
void ssl_process_exit(int pid);
void openssl_trace_handle(int pid, enum match_pids_act act);
void openssl_trace_init(void);
void set_uprobe_openssl_enabled(bool enabled);
bool is_openssl_trace_enabled(void);
#endif
