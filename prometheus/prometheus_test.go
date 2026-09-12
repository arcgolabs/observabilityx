package prometheus_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arcgolabs/observabilityx"
	promobs "github.com/arcgolabs/observabilityx/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestAdapterMetrics(t *testing.T) {
	t.Parallel()

	registry := prom.NewRegistry()
	obs := promobs.New(
		promobs.WithRegisterer(registry),
		promobs.WithGatherer(registry),
		promobs.WithNamespace("observabilityx_test"),
	)

	obs.Counter(observabilityx.NewCounterSpec("authx_authenticate_total", observabilityx.WithLabelKeys("result"))).
		Add(context.Background(), 1, observabilityx.String("result", "ok"))
	obs.UpDownCounter(observabilityx.NewUpDownCounterSpec("authx_inflight", observabilityx.WithLabelKeys("result"))).
		Add(context.Background(), 1, observabilityx.String("result", "ok"))
	obs.Histogram(observabilityx.NewHistogramSpec("authx_authenticate_duration_ms", observabilityx.WithLabelKeys("result"))).
		Record(context.Background(), 10, observabilityx.String("result", "ok"))
	obs.Gauge(observabilityx.NewGaugeSpec("authx_queue_depth", observabilityx.WithLabelKeys("result"))).
		Set(context.Background(), 3, observabilityx.String("result", "ok"))

	metrics, err := registry.Gather()
	require.NoError(t, err)
	require.NotEmpty(t, metrics)
}

func TestAdapterHandler(t *testing.T) {
	t.Parallel()

	registry := prom.NewRegistry()
	obs := promobs.New(promobs.WithRegisterer(registry), promobs.WithGatherer(registry))

	obs.Counter(observabilityx.NewCounterSpec("eventx_publish_total")).Add(context.Background(), 1)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", http.NoBody)
	w := httptest.NewRecorder()
	obs.Handler().ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "observabilityx_eventx_publish_total")
}

func TestAdaptersReuseAlreadyRegisteredCounter(t *testing.T) {
	t.Parallel()

	registry := prom.NewRegistry()
	first := promobs.New(promobs.WithRegisterer(registry))
	second := promobs.New(promobs.WithRegisterer(registry))
	spec := observabilityx.NewCounterSpec("shared_total")

	first.Counter(spec).Add(context.Background(), 1)
	second.Counter(spec).Add(context.Background(), 1)

	metrics, err := registry.Gather()
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Len(t, metrics[0].Metric, 1)
	require.Equal(t, float64(2), metrics[0].Metric[0].GetCounter().GetValue())
}

func TestAdaptersReuseAlreadyRegisteredCollectors(t *testing.T) {
	t.Parallel()

	registry := prom.NewRegistry()
	first := promobs.New(promobs.WithRegisterer(registry))
	second := promobs.New(promobs.WithRegisterer(registry))
	ctx := context.Background()

	first.Counter(observabilityx.NewCounterSpec("shared_counter_total")).Add(ctx, 1)
	second.Counter(observabilityx.NewCounterSpec("shared_counter_total")).Add(ctx, 1)
	first.UpDownCounter(observabilityx.NewUpDownCounterSpec("shared_up_down_total")).Add(ctx, 1)
	second.UpDownCounter(observabilityx.NewUpDownCounterSpec("shared_up_down_total")).Add(ctx, 1)
	first.Histogram(observabilityx.NewHistogramSpec("shared_histogram")).Record(ctx, 1)
	second.Histogram(observabilityx.NewHistogramSpec("shared_histogram")).Record(ctx, 1)
	first.Gauge(observabilityx.NewGaugeSpec("shared_gauge")).Set(ctx, 1)
	second.Gauge(observabilityx.NewGaugeSpec("shared_gauge")).Set(ctx, 2)

	metrics, err := registry.Gather()
	require.NoError(t, err)
	require.Len(t, metrics, 4)

	for _, metricFamily := range metrics {
		require.Len(t, metricFamily.Metric, 1)
		switch metricFamily.GetName() {
		case "observabilityx_shared_counter_total":
			require.Equal(t, float64(2), metricFamily.Metric[0].GetCounter().GetValue())
		case "observabilityx_shared_histogram":
			require.Equal(t, uint64(2), metricFamily.Metric[0].GetHistogram().GetSampleCount())
		case "observabilityx_shared_up_down_total", "observabilityx_shared_gauge":
			require.Equal(t, float64(2), metricFamily.Metric[0].GetGauge().GetValue())
		default:
			t.Fatalf("unexpected metric family %q", metricFamily.GetName())
		}
	}
}
