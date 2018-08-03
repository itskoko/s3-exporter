package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type bucketFlags []string

func (i *bucketFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *bucketFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var (
		listeningAddress = flag.String("telemetry.address", ":8080", "Address on which to expose metrics.")
		metricsEndpoint  = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")

		logger = log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), "caller", log.DefaultCaller)
	)
	buckets := bucketFlags{}
	flag.Var(&buckets, "b", "Bucket to get metrics for")
	flag.Parse()
	if len(buckets) == 0 {
		level.Error(logger).Log("msg", "You need to specify at least one bucket with -b")
		os.Exit(1)
	}

	session, err := session.NewSession()
	if err != nil {
		level.Error(logger).Log("msg", "Couldn't create AWS session", "error", err)
		os.Exit(1)
	}

	s3 := s3.New(session)
	exporter := NewExporter(logger, s3, buckets)
	prometheus.MustRegister(exporter)

	level.Info(logger).Log("msg", "Starting server", "address", *listeningAddress)
	http.Handle(*metricsEndpoint, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>S3 Exporter</title></head>
			<body>
			<h1>S3 Exporter</h1>
			<p><a href="` + *metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	if err := http.ListenAndServe(*listeningAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Server failed", "err", err.Error())
		os.Exit(1)
	}
}
