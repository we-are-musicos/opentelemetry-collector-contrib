// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translation // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter/internal/translation"

// DefaultExcludeMetricsYaml holds list of hard coded metrics that will added to the
// exclude list from the config. It includes non-default metrics collected by
// receivers. This list is determined by categorization of metrics in the SignalFx
// Agent. Metrics in the OpenTelemetry convention, that have equivalents in the
// SignalFx Agent that are categorized as non-default are also included in this list.
const DefaultExcludeMetricsYaml = `
exclude_metrics:

# Metrics in SignalFx Agent Format.
- metric_names:
  # CPU metrics.
  # Derived from https://docs.signalfx.com/en/latest/integrations/agent/monitors/cpu.html.
  - cpu.interrupt
  - cpu.nice
  - cpu.softirq
  - cpu.steal
  - cpu.system
  - cpu.user
  - cpu.utilization_per_core
  - cpu.wait

  # Disk-IO metrics.
  # Derived from https://docs.signalfx.com/en/latest/integrations/agent/monitors/disk-io.html.
  - disk_ops.pending

  # Virtual memory metrics
  - vmpage_io.memory.in
  - vmpage_io.memory.out


# Metrics in OpenTelemetry Convention.

# CPU Metrics.
- metric_name: system.cpu.time
  dimensions:
    state: [interrupt, nice, softirq, steal, system, user, wait]

- metric_name: cpu.idle
  dimensions:
    cpu: ["*"]

# Memory metrics.
- metric_name: system.memory.usage
  dimensions:
    state: [inactive]

# Filesystem metrics.
- metric_name: system.filesystem.usage
  dimensions:
    state: [reserved]
- metric_name: system.filesystem.inodes.usage

# Disk-IO metrics.
- metric_names:
  - system.disk.merged
  - system.disk.io
  - system.disk.time
  - system.disk.io_time
  - system.disk.operation_time
  - system.disk.pending_operations
  - system.disk.weighted_io_time

# Network-IO metrics.
- metric_names:
  - system.network.packets
  - system.network.dropped
  - system.network.tcp_connections
  - system.network.connections

# Processes metrics
- metric_names:
  - system.processes.count
  - system.processes.created

# Virtual memory metrics.
- metric_names:
  - system.paging.faults
  - system.paging.usage
- metric_name: system.paging.operations
  dimensions:
    type: [minor]

# k8s metrics.
- metric_names:
  - k8s.cronjob.active_jobs
  - k8s.job.active_pods
  - k8s.job.desired_successful_pods
  - k8s.job.failed_pods
  - k8s.job.max_parallel_pods
  - k8s.job.successful_pods
  - k8s.statefulset.desired_pods
  - k8s.statefulset.current_pods
  - k8s.statefulset.ready_pods
  - k8s.statefulset.updated_pods
  - k8s.hpa.max_replicas
  - k8s.hpa.min_replicas
  - k8s.hpa.current_replicas
  - k8s.hpa.desired_replicas

  # matches all container limit metrics but k8s.container.cpu_limit and k8s.container.memory_limit
  - /^k8s\.container\..+_limit$/
  - '!k8s.container.memory_limit'
  - '!k8s.container.cpu_limit'

  # matches all container request metrics but k8s.container.cpu_request and k8s.container.memory_request
  - /^k8s\.container\..+_request$/
  - '!k8s.container.memory_request'
  - '!k8s.container.cpu_request'

  # matches any node condition but memory_pressure, network_unavailable, out_of_disk, p_i_d_pressure, and ready
  - /^k8s\.node\.condition_.+$/
  - '!k8s.node.condition_memory_pressure'
  - '!k8s.node.condition_network_unavailable'
  - '!k8s.node.condition_out_of_disk'
  - '!k8s.node.condition_p_i_d_pressure'
  - '!k8s.node.condition_ready'

  # kubelet metrics
  # matches (container|k8s.node|k8s.pod).memory...
  - /^(?i:(container)|(k8s\.node)|(k8s\.pod))\.memory\.available$/
  - /^(?i:(container)|(k8s\.node)|(k8s\.pod))\.memory\.major_page_faults$/
  - /^(?i:(container)|(k8s\.node)|(k8s\.pod))\.memory\.page_faults$/
  - /^(?i:(container)|(k8s\.node)|(k8s\.pod))\.memory\.rss$/
  - /^(?i:(k8s\.node)|(k8s\.pod))\.memory\.usage$/
  - /^(?i:(container)|(k8s\.node)|(k8s\.pod))\.memory\.working_set$/

  # matches (k8s.node|k8s.pod).filesystem...
  - /^k8s\.(?i:(node)|(pod))\.filesystem\.available$/
  - /^k8s\.(?i:(node)|(pod))\.filesystem\.capacity$/
  - /^k8s\.(?i:(node)|(pod))\.filesystem\.usage$/

  # matches (k8s.node|k8s.pod).cpu.time
  - /^k8s\.(?i:(node)|(pod))\.cpu\.time$/

  # matches (k8s.node|k8s.pod).cpu.utilization
  - /^(?i:(k8s\.node)|(k8s\.pod))\.cpu\.utilization$/

  # matches k8s.node.network.io and k8s.node.network.errors
  - /^k8s\.node\.network\.(?:(io)|(errors))$/

  # matches k8s.volume.inodes, k8s.volume.inodes and k8s.volume.inodes.used
  - /^k8s\.volume\.inodes(\.free|\.used)*$/
`
