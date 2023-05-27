// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

// Package metrics provides standard READs metrics for HTTP applications, as well as a way to define and set
// custom metrics.
//
// This package automatically creates three metrics on initialization:
//  1. http_requests_total: A counter metric, which counts all http requests (assuming you include the [GinMiddleware]
//     in your project).
//     The status and path of the request are included as labels
//  2. http_errors_total: A counter metric, which counts all http requests which have a status < 200 or > 299
//     The status and path of the request are included as labels
//  3. http_request_duration_seconds: A histogram which tracks the duration of each request
//     The status and path of the request are included as labels
package metrics

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	statusLabel = "status"
	routeLabel  = "route"

	requestDurationKey = "http_request_duration_seconds"
	totalRequestsKey   = "http_requests_total"
	totalErrorsKey     = "http_errors_total"
)

type counterMetric struct {
	key     string
	counter *prometheus.CounterVec
	labels  []string
}

func (m *counterMetric) With(labels map[string]string) prometheus.Counter {
	return m.counter.With(unifyLabels(m.labels, labels))
}

type histogramMetric struct {
	key       string
	histogram *prometheus.HistogramVec
	labels    []string
}

func (m *histogramMetric) With(labels map[string]string) prometheus.Observer {
	return m.histogram.With(unifyLabels(m.labels, labels))
}

var (
	registry       = prometheus.DefaultRegisterer
	counters       = make(map[string]*counterMetric, 0)
	histograms     = make(map[string]*histogramMetric, 0)
	defaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 15.0, 60.0, 300.0, 1200.0, math.MaxFloat64}
)

func getCounter(name string) *counterMetric {
	val, ok := counters[name]
	if ok {
		return val
	}
	return nil
}

func getHistogram(name string) *histogramMetric {
	val, ok := histograms[name]
	if ok {
		return val
	}
	return nil
}

func unifyLabels(keyList []string, labels map[string]string) map[string]string {
	finalMap := make(map[string]string)
	for _, key := range keyList {
		v, ok := labels[key]
		if ok {
			finalMap[key] = v
		} else {
			finalMap[key] = ""
		}
	}
	return finalMap
}

// AddCounter adds a new counter metric
// A new counter metric is added with the name provided by key, and a description provided by help.
// The labels provided should be any label that _could_ be associated with the counter,
// even if it's not _always_ associated with the counter
//
// See the [counter docs] for more information
//
// [counter docs]: https://prometheus.io/docs/tutorials/understanding_metric_types/#counter
func AddCounter(key string, help string, labels []string) {
	_, ok := histograms[key]
	if ok {
		return
	}
	newCounter := &counterMetric{
		key:    key,
		labels: labels,
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: key,
				Help: help,
			},
			labels,
		),
	}
	counters[key] = newCounter
	registry.MustRegister(newCounter.counter)
}

// AddHistogram adds a new histogram metric
// A new histogram metric is added with the name provided by key, and a description provided by help.
// The labels provided should be any label that _could_ be associated with the histogram,
// even if it's not _always_ associated with the histogram.
//
// All histograms are created with a standard set of buckets which are:
//
//	[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 15.0, 60.0, 300.0, 1200.0, +Infiinity]
//
// See the [histogram docs] for more information
//
// [histogram docs]: https://prometheus.io/docs/tutorials/understanding_metric_types/#histogram
func AddHistogram(key string, help string, labels []string) {
	_, ok := counters[key]
	if ok {
		return
	}
	newHistogram := &histogramMetric{
		key:    key,
		labels: labels,
		histogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    key,
				Help:    help,
				Buckets: defaultBuckets,
			},
			labels,
		),
	}
	histograms[key] = newHistogram
	registry.MustRegister(newHistogram.histogram)
}

// IncrementCounter will increment the counter specified by key by 1.
// The labels provided by the labels property will be compared to the list of labels used to create the counter.
// Any labels _not_ included in this call that were specified when the counter was created
// will be given blank values (empty strings).
// Any labels included in this call which were _not_ included when the counter was created
// will be silently ignored
func IncrementCounter(key string, labels map[string]string) {
	counter := getCounter(key)
	if counter == nil {
		return
	}
	counter.With(labels).Inc()
}

// UpdateHistogram will update the histogram specified by key with the duration from howLong as Seconds.
// So a duration of 500ms will be recorded as 0.5 in the histogram.
// The labels provided by the labels property will be compared to the list of labels used to create the histogram.
// Any labels _not_ included in this call there were specified when the histogram was created
// will be given blank values (empty strings).
// Any labels included in this call which were _not_ included when the histogram was created
// will be silently ignored
func UpdateHistogram(key string, howLong time.Duration, labels map[string]string) {
	histogram := getHistogram(key)
	if histogram == nil {
		return
	}
	histogram.With(labels).Observe(howLong.Seconds())
}

func init() {
	AddCounter(totalRequestsKey, "Total number of HTTP requests", []string{statusLabel, routeLabel})
	AddCounter(totalErrorsKey, "Total number of errors (a status other than 2XX)", []string{statusLabel, routeLabel})
	AddHistogram(requestDurationKey, "Duration of all HTTP requests", []string{statusLabel, routeLabel})
}

// GinMiddleware is a [gin.HandlerFunc] which will update the default http_requests_total, http_errors_total,
// and http_request_duration_seconds for each request.
// To include it in your service, just pass it to the `Use` function like so:
//
//	r := gin.Default()
//	r.Use(metrics.GinMiddleware)
func GinMiddleware(c *gin.Context) {
	path := c.Request.URL.Path
	s := time.Now()
	c.Next()
	e := time.Since(s)
	status := c.Writer.Status()
	statusLabels := map[string]string{statusLabel: fmt.Sprintf("%d", status), routeLabel: path}
	UpdateHistogram(requestDurationKey, e, statusLabels)
	IncrementCounter(totalRequestsKey, statusLabels)
	if status < 200 || status > 299 {
		IncrementCounter(totalErrorsKey, statusLabels)
	}
}

// GinMetricsHandler is a [gin.HandlerFunc] which will expose all the metric data at a specified endpoint,
// including standard Go metrics provided by the Prometheus client.
//
// Adding this handler is required for services to have their metrics shipped to Funnel and Argus.
// To add it to your service, simply create a GET endpoint, and provide the handler like so:
//
//		r := gin.Default()
//	    r.GET("/__metrics", metrics.GinMetricsHandler)
func GinMetricsHandler(c *gin.Context) {
	gin.WrapH(promhttp.Handler())(c)
}
