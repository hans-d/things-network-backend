// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package stats supports the collection of metrics from a running component.
package stats

import "github.com/rcrowley/go-metrics"

// MarkMeter registers an event
func MarkMeter(name string) {
	metrics.GetOrRegisterMeter(name, metrics.DefaultRegistry).Mark(1)
}

// UpdateHistogram registers a new value for a histogram
func UpdateHistogram(name string, value int64) {
	metrics.GetOrRegisterHistogram(name, metrics.DefaultRegistry, metrics.NewUniformSample(1000)).Update(value)
}

// IncCounter increments a counter by 1
func IncCounter(name string) {
	metrics.GetOrRegisterCounter(name, metrics.DefaultRegistry).Inc(1)
}

// DecCounter decrements a counter by 1
func DecCounter(name string) {
	metrics.GetOrRegisterCounter(name, metrics.DefaultRegistry).Dec(1)
}
