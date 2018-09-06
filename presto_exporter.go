package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
)

const (
	namespace = "presto_cluster"
)

type Exporter struct {
	uri              string
	RunningQueries   float64 `json:"runningQueries"`
	BlockedQueries   float64 `json:"blockedQueries"`
	QueuedQueries    float64 `json:"queuedQueries"`
	ActiveWorkers    float64 `json:"activeWorkers"`
	RunningDrivers   float64 `json:"runningDrivers"`
	ReservedMemory   float64 `json:"reservedMemory"`
	TotalInputRows   float64 `json:"totalInputRows"`
	TotalInputBytes  float64 `json:"totalInputBytes"`
	TotalCpuTimeSecs float64 `json:"totalCpuTimeSecs"`
}

var (
	runningQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "running_queries"),
		"Running requests of the presto cluster.",
		nil, nil,
	)
	blockedQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "blocked_queries"),
		"Blocked queries of the presto cluster.",
		nil, nil,
	)
	queuedQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "queued_queries"),
		"Queued queries of the presto cluster.",
		nil, nil,
	)
	activeWorkers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "active_workers"),
		"Active workers of the presto cluster.",
		nil, nil,
	)
	runningDrivers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "running_drivers"),
		"Running drivers of the presto cluster.",
		nil, nil,
	)
	reservedMemory = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "reserved_memory"),
		"Reserved memory of the presto cluster.",
		nil, nil,
	)
	totalInputRows = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_input_rows"),
		"Total input rows of the presto cluster.",
		nil, nil,
	)
	totalInputBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_input_bytes"),
		"Total input bytes of the presto cluster.",
		nil, nil,
	)
	totalCpuTimeSecs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_cpu_time_secs"),
		"Total cpu time of the presto cluster.",
		nil, nil,
	)
)

// Describe implements the prometheus.Collector interface.
func (e Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- runningQueries
	ch <- blockedQueries
	ch <- queuedQueries
	ch <- activeWorkers
	ch <- runningDrivers
	ch <- reservedMemory
	ch <- totalInputRows
	ch <- totalInputBytes
	ch <- totalCpuTimeSecs
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9482").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		opts          = Exporter{}
	)
	kingpin.Flag("web.url", "Presto cluster address.").Default("http://localhost:8080/v1/cluster").StringVar(&opts.uri)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("presto_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting presto_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(&Exporter{uri: opts.uri})

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Presto Exporter</title></head>
			<body>
  		<h1>Presto Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
  		</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	resp, err := http.Get(e.uri)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	if resp.StatusCode != 200 {
		log.Errorf("%s", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	err = json.Unmarshal(body, &e)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(runningQueries, prometheus.GaugeValue, e.RunningQueries)
	ch <- prometheus.MustNewConstMetric(blockedQueries, prometheus.GaugeValue, e.BlockedQueries)
	ch <- prometheus.MustNewConstMetric(queuedQueries, prometheus.GaugeValue, e.QueuedQueries)
	ch <- prometheus.MustNewConstMetric(activeWorkers, prometheus.GaugeValue, e.ActiveWorkers)
	ch <- prometheus.MustNewConstMetric(runningDrivers, prometheus.GaugeValue, e.RunningDrivers)
	ch <- prometheus.MustNewConstMetric(reservedMemory, prometheus.GaugeValue, e.ReservedMemory)
	ch <- prometheus.MustNewConstMetric(totalInputRows, prometheus.GaugeValue, e.TotalInputRows)
	ch <- prometheus.MustNewConstMetric(totalInputBytes, prometheus.GaugeValue, e.TotalInputBytes)
	ch <- prometheus.MustNewConstMetric(totalCpuTimeSecs, prometheus.GaugeValue, e.TotalCpuTimeSecs)
}
