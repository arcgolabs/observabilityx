package observabilityx

import (
	"context"
	"log/slog"

	collectionlist "github.com/arcgolabs/collectionx/list"
)

// Multi combines multiple observability backends into one.
//
// Use this to send telemetry to more than one backend (for example OTel + Prometheus).
func Multi(backends ...Observability) Observability {
	filtered := collectionlist.RejectList(collectionlist.NewList(backends...), func(_ int, backend Observability) bool {
		return backend == nil
	})
	if filtered.IsEmpty() {
		return Nop()
	}

	firstBackend, _ := filtered.GetFirst()
	logger := firstBackend.Logger()
	if logger == nil {
		logger = slog.Default()
	}

	return &multiObservability{
		backends: filtered,
		logger:   logger,
	}
}

type multiObservability struct {
	backends *collectionlist.List[Observability]
	logger   *slog.Logger
}

func (m *multiObservability) Logger() *slog.Logger {
	return NormalizeLogger(m.logger)
}

func (m *multiObservability) StartSpan(
	ctx context.Context,
	name string,
	attrs ...Attribute,
) (context.Context, Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	firstBackend, _ := m.backends.GetFirst()
	nextCtx, firstSpan := firstBackend.StartSpan(ctx, name, attrs...)
	spans := collectionlist.NewListWithCapacity[Span](m.backends.Len())
	if firstSpan != nil {
		spans.Add(firstSpan)
	}
	m.backends.Range(func(index int, backend Observability) bool {
		if index == 0 {
			return true
		}
		_, span := backend.StartSpan(nextCtx, name, attrs...)
		if span != nil {
			spans.Add(span)
		}
		return true
	})
	if spans.Len() == 0 {
		return nextCtx, nopSpan{}
	}
	return nextCtx, multiSpan{spans: spans}
}

func (m *multiObservability) Counter(spec CounterSpec) Counter {
	counters := m.collect(func(backend Observability) Counter {
		return backend.Counter(spec)
	})
	if counters.IsEmpty() {
		return nopCounter{}
	}
	return multiCounter{counters: counters}
}

func (m *multiObservability) UpDownCounter(spec UpDownCounterSpec) UpDownCounter {
	counters := m.collect(func(backend Observability) UpDownCounter {
		return backend.UpDownCounter(spec)
	})
	if counters.IsEmpty() {
		return nopUpDownCounter{}
	}
	return multiUpDownCounter{counters: counters}
}

func (m *multiObservability) Histogram(spec HistogramSpec) Histogram {
	histograms := m.collect(func(backend Observability) Histogram {
		return backend.Histogram(spec)
	})
	if histograms.IsEmpty() {
		return nopHistogram{}
	}
	return multiHistogram{histograms: histograms}
}

func (m *multiObservability) Gauge(spec GaugeSpec) Gauge {
	gauges := m.collect(func(backend Observability) Gauge {
		return backend.Gauge(spec)
	})
	if gauges.IsEmpty() {
		return nopGauge{}
	}
	return multiGauge{gauges: gauges}
}

// collect builds one typed handle per backend and drops nil handles.
// Go 1.27 generic methods let the fan-out logic stay shared while preserving
// the concrete handle type at each public metric method.
func (m *multiObservability) collect[T any](build func(Observability) T) *collectionlist.List[T] {
	handles := collectionlist.NewListWithCapacity[T](m.backends.Len())
	m.backends.Range(func(_ int, backend Observability) bool {
		handle := build(backend)
		if any(handle) != nil {
			handles.Add(handle)
		}
		return true
	})
	return handles
}

type multiCounter struct {
	counters *collectionlist.List[Counter]
}

func (m multiCounter) Add(ctx context.Context, value int64, attrs ...Attribute) {
	m.counters.Range(func(_ int, counter Counter) bool {
		counter.Add(ctx, value, attrs...)
		return true
	})
}

type multiUpDownCounter struct {
	counters *collectionlist.List[UpDownCounter]
}

func (m multiUpDownCounter) Add(ctx context.Context, value int64, attrs ...Attribute) {
	m.counters.Range(func(_ int, counter UpDownCounter) bool {
		counter.Add(ctx, value, attrs...)
		return true
	})
}

type multiHistogram struct {
	histograms *collectionlist.List[Histogram]
}

func (m multiHistogram) Record(ctx context.Context, value float64, attrs ...Attribute) {
	m.histograms.Range(func(_ int, histogram Histogram) bool {
		histogram.Record(ctx, value, attrs...)
		return true
	})
}

type multiGauge struct {
	gauges *collectionlist.List[Gauge]
}

func (m multiGauge) Set(ctx context.Context, value float64, attrs ...Attribute) {
	m.gauges.Range(func(_ int, gauge Gauge) bool {
		gauge.Set(ctx, value, attrs...)
		return true
	})
}

type multiSpan struct {
	spans *collectionlist.List[Span]
}

func (s multiSpan) End() {
	s.spans.Range(func(_ int, span Span) bool {
		span.End()
		return true
	})
}

func (s multiSpan) RecordError(err error) {
	if err == nil {
		return
	}
	s.spans.Range(func(_ int, span Span) bool {
		span.RecordError(err)
		return true
	})
}

func (s multiSpan) SetAttributes(attrs ...Attribute) {
	s.spans.Range(func(_ int, span Span) bool {
		span.SetAttributes(attrs...)
		return true
	})
}
