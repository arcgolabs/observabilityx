package prometheus

import (
	"testing"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestAsAlreadyRegisteredErrorSupportsValueAndPointer(t *testing.T) {
	t.Parallel()

	collector := prom.NewCounterVec(prom.CounterOpts{Name: "test_total"}, nil)
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "value",
			err:  prom.AlreadyRegisteredError{ExistingCollector: collector},
		},
		{
			name: "pointer",
			err:  &prom.AlreadyRegisteredError{ExistingCollector: collector},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := asAlreadyRegisteredError(tt.err)
			require.True(t, ok)
			require.Same(t, collector, got.ExistingCollector)
		})
	}
}
