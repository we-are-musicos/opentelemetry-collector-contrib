# Windows Performance Counters Receiver

This receiver, for Windows only, captures the configured system, application, or
custom performance counter data from the Windows registry using the [PDH
interface](https://docs.microsoft.com/en-us/windows/win32/perfctrs/using-the-pdh-functions-to-consume-counter-data).
It is based on the [Telegraf Windows Performance Counters Input
Plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/win_perf_counters).

- `Memory\Committed Bytes`
- `Processor\% Processor Time`, with a datapoint for each `Instance` label = (`_Total`, `1`, `2`, `3`, ... )

If one of the specified performance counters cannot be loaded on startup, a
warning will be printed, but the application will not fail fast. It is expected
that some performance counters may not exist on some systems due to different OS
configuration.

## Configuration

The collection interval and the list of performance counters to be scraped can
be configured:

```yaml
windowsperfcounters:
  collection_interval: <duration> # default = "1m"
  metrics:
    <metric name>:
      description: <description>
      unit: <unit type>
      gauge:
    <metric name>:
      description: <description>
      unit: <unit type>
      sum:
        aggregation: <cumulative or delta>
        monotonic: <true or false>
  perfcounters:
    - object: <object name>
      instances: [<instance name>]*
      counters:
        - name: <counter name>
          metric: <metric name>
          attributes:
            <key>: <value>
```

*Note `instances` can have several special values depending on the type of
counter:

Value | Interpretation
-- | --
Not specified | This is the only valid value if the counter has no instances
`"*"` | All instances
`"_Total"` | The "total" instance
`"instance1"` | A single instance
`["instance1", "instance2", ...]` | A set of instances
`["_Total", "instance1", "instance2", ...]` | A set of instances including the "total" instance

### Scraping at different frequencies

If you would like to scrape some counters at a different frequency than others,
you can configure multiple `windowsperfcounters` receivers with different
`collection_interval` values. For example:

```yaml
receivers:
  windowsperfcounters/memory:
    metrics:
      bytes.committed:
        description: the number of bytes committed to memory
        unit: By
        gauge:
    collection_interval: 30s
    perfcounters:
      - object: Memory
        counters:
          - name: Committed Bytes
            metric: bytes.committed

  windowsperfcounters/processor:
    collection_interval: 1m
    metrics:
      processor.time:
        description: active and idle time of the processor
        unit: "%"
        gauge:
    perfcounters:
      - object: "Processor"
        instances: "*"
        counters:
          - name: "% Processor Time"
            metric: processor.time
            attributes:
              state: active
      - object: "Processor"
        instances: [1, 2]
        counters:
          - name: "% Idle Time"
            metric: processor.time
            attributes:
              state: idle

service:
  pipelines:
    metrics:
      receivers: [windowsperfcounters/memory, windowsperfcounters/processor]
```

### Defining metric format

To report metrics in the desired output format, define a metric and reference it in the corresponding counter, along with any applicable attributes. The metric's data type can either be `gauge` (default) or `sum`. 

| Field Name  | Description                              | Value        | Default      |
| --          | --                                       | --           | --           |
| name        | The key for the metric.                  | string       | Counter Name |
| description | definition of what the metric measures.  | string       |              |
| unit        | what is being measured.                  | string       | `1`          |
| sum         | representation of a sum metric.          | Sum Config   |              |
| gauge       | representation of a gauge metric.        | Gauge Config |              |


#### Sum Config

| Field Name   | Description                                           | Value                           | Default |
| --           | --                                                    | --                              | --      |
| aggregation  | The type of aggregation temporality for the metric.   | [`cumulative` or `delta`]       |         |
| monotonic    | whether or not the metric value can decrease.         | false                           |         |

#### Gauge Config

A `gauge` config currently accepts no settings. It is specified as an object for forwards compatibility.

e.g. To output the `Memory/Committed Bytes` counter as a metric with the name
`bytes.committed`:

```yaml
receivers:
  windowsperfcounters:
    metrics:
      bytes.committed:
        description: the number of bytes committed to memory
        unit: By
        gauge:
    collection_interval: 30s
    perfcounters:
    - object: Memory
      counters:
        - name: Committed Bytes
          metric: bytes.committed

service:
  pipelines:
    metrics:
      receivers: [windowsperfcounters]
```

## Known Limitation
- The network interface is not available inside the container. Hence, the metrics for the object `Network Interface` aren't generated in that scenario. In the case of sub-process, it captures `Network Interface` metrics. There is a similar open issue in [Github](https://github.com/influxdata/telegraf/issues/5357) and [Docker](https://forums.docker.com/t/unable-to-collect-network-metrics-inside-windows-container-on-windows-server-2016-data-center/69480) forum.
