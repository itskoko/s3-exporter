package main

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "s3" // For Prometheus metrics.
)

// Exporter collects nginx stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	logger log.Logger
	s3     s3iface.S3API

	buckets         []string
	errorCounter    prometheus.Counter
	bucketItemCount *prometheus.Desc
}

// NewExporter returns an initialized Exporter.
func NewExporter(logger log.Logger, s3 s3iface.S3API, buckets []string) *Exporter {
	return &Exporter{
		logger:  logger,
		s3:      s3,
		buckets: buckets,
		errorCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Total number of errors",
		}),
		bucketItemCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bucket_item_count"),
			"Number of items in given bucket",
			[]string{"bucket"},
			nil,
		),
	}
}

// Describe describes all the metrics ever exported by the nginx exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.errorCounter.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, bucket := range e.buckets {
		n, err := e.countObjects(bucket)
		if err != nil {
			level.Error(e.logger).Log("msg", "Couldn't list objects", "err", err.Error(), "bucket", bucket)
			e.errorCounter.Inc()
			continue
		}
		ch <- prometheus.MustNewConstMetric(e.bucketItemCount, prometheus.GaugeValue, float64(n), bucket)
	}
	ch <- e.errorCounter
}

func (e *Exporter) countObjects(bucket string) (int, error) {
	n := 0
	ctx := context.Background()
	err := e.s3.ListObjectsPagesWithContext(ctx,
		&s3.ListObjectsInput{
			Bucket: &bucket,
		},
		func(page *s3.ListObjectsOutput, lastPage bool) bool {
			n += len(page.Contents)
			return true
		},
	)
	return n, err
}
