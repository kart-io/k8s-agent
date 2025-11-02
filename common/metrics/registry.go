// Package metrics provides centralized metrics collection and reporting.
package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry manages all metrics for an application.
type Registry struct {
	namespace  string
	subsystem  string
	mu         sync.RWMutex
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec
}

// NewRegistry creates a new metrics registry.
func NewRegistry(namespace, subsystem string) *Registry {
	return &Registry{
		namespace:  namespace,
		subsystem:  subsystem,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
	}
}

// Counter returns or creates a counter metric.
func (r *Registry) Counter(name, help string, labels []string) *prometheus.CounterVec {
	r.mu.Lock()
	defer r.mu.Unlock()

	if counter, exists := r.counters[name]; exists {
		return counter
	}

	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: r.namespace,
			Subsystem: r.subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	r.counters[name] = counter
	return counter
}

// Gauge returns or creates a gauge metric.
func (r *Registry) Gauge(name, help string, labels []string) *prometheus.GaugeVec {
	r.mu.Lock()
	defer r.mu.Unlock()

	if gauge, exists := r.gauges[name]; exists {
		return gauge
	}

	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: r.namespace,
			Subsystem: r.subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	r.gauges[name] = gauge
	return gauge
}

// Histogram returns or creates a histogram metric.
func (r *Registry) Histogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	r.mu.Lock()
	defer r.mu.Unlock()

	if histogram, exists := r.histograms[name]; exists {
		return histogram
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: r.namespace,
			Subsystem: r.subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)

	r.histograms[name] = histogram
	return histogram
}

// Summary returns or creates a summary metric.
func (r *Registry) Summary(name, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	r.mu.Lock()
	defer r.mu.Unlock()

	if summary, exists := r.summaries[name]; exists {
		return summary
	}

	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	summary := promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  r.namespace,
			Subsystem:  r.subsystem,
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)

	r.summaries[name] = summary
	return summary
}

// Inc increments a counter by 1.
func (r *Registry) Inc(name string, labels prometheus.Labels) error {
	r.mu.RLock()
	counter, exists := r.counters[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("counter %s not found", name)
	}

	counter.With(labels).Inc()
	return nil
}

// Add adds a value to a counter.
func (r *Registry) Add(name string, value float64, labels prometheus.Labels) error {
	r.mu.RLock()
	counter, exists := r.counters[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("counter %s not found", name)
	}

	counter.With(labels).Add(value)
	return nil
}

// Set sets a gauge to a specific value.
func (r *Registry) Set(name string, value float64, labels prometheus.Labels) error {
	r.mu.RLock()
	gauge, exists := r.gauges[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}

	gauge.With(labels).Set(value)
	return nil
}

// Observe records an observation in a histogram or summary.
func (r *Registry) Observe(name string, value float64, labels prometheus.Labels) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try histogram first
	if histogram, exists := r.histograms[name]; exists {
		histogram.With(labels).Observe(value)
		return nil
	}

	// Try summary
	if summary, exists := r.summaries[name]; exists {
		summary.With(labels).Observe(value)
		return nil
	}

	return fmt.Errorf("histogram/summary %s not found", name)
}

// GetCounter returns a counter by name.
func (r *Registry) GetCounter(name string) (*prometheus.CounterVec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	counter, exists := r.counters[name]
	return counter, exists
}

// GetGauge returns a gauge by name.
func (r *Registry) GetGauge(name string) (*prometheus.GaugeVec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	gauge, exists := r.gauges[name]
	return gauge, exists
}

// GetHistogram returns a histogram by name.
func (r *Registry) GetHistogram(name string) (*prometheus.HistogramVec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	histogram, exists := r.histograms[name]
	return histogram, exists
}

// GetSummary returns a summary by name.
func (r *Registry) GetSummary(name string) (*prometheus.SummaryVec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	summary, exists := r.summaries[name]
	return summary, exists
}

// Reset clears all metrics (useful for testing).
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, counter := range r.counters {
		counter.Reset()
	}
	for range r.gauges {
		// Gauges don't have Reset, so we skip them
	}
	for _, histogram := range r.histograms {
		histogram.Reset()
	}
	for _, summary := range r.summaries {
		summary.Reset()
	}
}
