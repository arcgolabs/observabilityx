package observabilityx

import (
	"strings"
	"unicode"

	collectionlist "github.com/arcgolabs/collectionx/list"
	"github.com/arcgolabs/pkg/option"
	"github.com/samber/lo"
)

// MetricOption configures shared metric spec fields.
type MetricOption func(*MetricSpec)

// MetricSpec contains shared metric declaration fields.
type MetricSpec struct {
	Name        string
	Description string
	Unit        string
	LabelKeys   *collectionlist.List[string]
}

// CounterSpec declares an increasing int64 metric.
type CounterSpec struct {
	MetricSpec
}

// UpDownCounterSpec declares a signed int64 metric.
type UpDownCounterSpec struct {
	MetricSpec
}

// HistogramSpec declares a float64 histogram metric.
type HistogramSpec struct {
	MetricSpec
	Buckets *collectionlist.List[float64]
}

// GaugeSpec declares a float64 gauge metric.
type GaugeSpec struct {
	MetricSpec
}

// WithDescription sets a metric description.
func WithDescription(description string) MetricOption {
	return func(spec *MetricSpec) {
		spec.Description = description
	}
}

// WithUnit sets a metric unit.
func WithUnit(unit string) MetricOption {
	return func(spec *MetricSpec) {
		spec.Unit = unit
	}
}

// WithLabelKeys sets the declared metric label keys.
func WithLabelKeys(labelKeys ...string) MetricOption {
	return func(spec *MetricSpec) {
		spec.LabelKeys = collectionlist.NewList(labelKeys...)
	}
}

// WithMetricLabels sets the declared metric label keys from a collectionx/list value.
func WithMetricLabels(labelKeys *collectionlist.List[string]) MetricOption {
	return func(spec *MetricSpec) {
		spec.LabelKeys = cloneStringList(labelKeys)
	}
}

// NewCounterSpec creates a normalized counter spec.
func NewCounterSpec(name string, opts ...MetricOption) CounterSpec {
	spec := CounterSpec{
		MetricSpec: MetricSpec{
			Name:      name,
			LabelKeys: collectionlist.NewList[string](),
		},
	}
	option.Apply(&spec.MetricSpec, opts...)
	return NormalizeCounterSpec(spec)
}

// NewUpDownCounterSpec creates a normalized up-down counter spec.
func NewUpDownCounterSpec(name string, opts ...MetricOption) UpDownCounterSpec {
	spec := UpDownCounterSpec{
		MetricSpec: MetricSpec{
			Name:      name,
			LabelKeys: collectionlist.NewList[string](),
		},
	}
	option.Apply(&spec.MetricSpec, opts...)
	return NormalizeUpDownCounterSpec(spec)
}

// NewHistogramSpec creates a normalized histogram spec.
func NewHistogramSpec(name string, opts ...MetricOption) HistogramSpec {
	spec := HistogramSpec{
		MetricSpec: MetricSpec{
			Name:      name,
			LabelKeys: collectionlist.NewList[string](),
		},
		Buckets: collectionlist.NewList[float64](),
	}
	option.Apply(&spec.MetricSpec, opts...)
	return NormalizeHistogramSpec(spec)
}

// NewGaugeSpec creates a normalized gauge spec.
func NewGaugeSpec(name string, opts ...MetricOption) GaugeSpec {
	spec := GaugeSpec{
		MetricSpec: MetricSpec{
			Name:      name,
			LabelKeys: collectionlist.NewList[string](),
		},
	}
	option.Apply(&spec.MetricSpec, opts...)
	return NormalizeGaugeSpec(spec)
}

// NormalizeMetricSpec returns a normalized metric spec copy.
func NormalizeMetricSpec(spec MetricSpec) MetricSpec {
	return MetricSpec{
		Name:        strings.TrimSpace(spec.Name),
		Description: strings.TrimSpace(spec.Description),
		Unit:        strings.TrimSpace(spec.Unit),
		LabelKeys:   normalizeLabelKeys(spec.LabelKeys),
	}
}

// NormalizeCounterSpec returns a normalized counter spec copy.
func NormalizeCounterSpec(spec CounterSpec) CounterSpec {
	spec.MetricSpec = NormalizeMetricSpec(spec.MetricSpec)
	return spec
}

// NormalizeUpDownCounterSpec returns a normalized up-down counter spec copy.
func NormalizeUpDownCounterSpec(spec UpDownCounterSpec) UpDownCounterSpec {
	spec.MetricSpec = NormalizeMetricSpec(spec.MetricSpec)
	return spec
}

// NormalizeHistogramSpec returns a normalized histogram spec copy.
func NormalizeHistogramSpec(spec HistogramSpec) HistogramSpec {
	spec.MetricSpec = NormalizeMetricSpec(spec.MetricSpec)
	spec.Buckets = normalizeBuckets(spec.Buckets)
	return spec
}

// WithBuckets returns a normalized histogram spec copy with custom buckets.
func (spec HistogramSpec) WithBuckets(buckets ...float64) HistogramSpec {
	spec.Buckets = collectionlist.NewList(buckets...)
	return NormalizeHistogramSpec(spec)
}

// WithBucketList returns a normalized histogram spec copy with custom buckets.
func (spec HistogramSpec) WithBucketList(buckets *collectionlist.List[float64]) HistogramSpec {
	spec.Buckets = cloneFloat64List(buckets)
	return NormalizeHistogramSpec(spec)
}

// NormalizeGaugeSpec returns a normalized gauge spec copy.
func NormalizeGaugeSpec(spec GaugeSpec) GaugeSpec {
	spec.MetricSpec = NormalizeMetricSpec(spec.MetricSpec)
	return spec
}

func normalizeLabelKeys(labelKeys *collectionlist.List[string]) *collectionlist.List[string] {
	values := lo.FilterMap(valuesOrEmpty(labelKeys), func(labelKey string, _ int) (string, bool) {
		normalized := normalizeMetricLabelKey(labelKey)
		return normalized, normalized != ""
	})
	return collectionlist.NewList(lo.Uniq(values)...)
}

func normalizeBuckets(buckets *collectionlist.List[float64]) *collectionlist.List[float64] {
	values := lo.Filter(valuesOrEmpty(buckets), func(bucket float64, _ int) bool {
		return bucket > 0
	})
	return collectionlist.NewList(values...)
}

func cloneStringList(values *collectionlist.List[string]) *collectionlist.List[string] {
	return collectionlist.NewList(valuesOrEmpty(values)...)
}

func cloneFloat64List(values *collectionlist.List[float64]) *collectionlist.List[float64] {
	return collectionlist.NewList(valuesOrEmpty(values)...)
}

func valuesOrEmpty[T any](values *collectionlist.List[T]) []T {
	if values == nil {
		return nil
	}
	return values.Values()
}

func normalizeMetricLabelKey(raw string) string {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return ""
	}

	replaced := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
			return unicode.ToLower(r)
		default:
			return '_'
		}
	}, clean)
	replaced = strings.Trim(replaced, "_")
	if replaced == "" {
		return ""
	}

	firstRune := rune(replaced[0])
	if !unicode.IsLetter(firstRune) && firstRune != '_' {
		replaced = "_" + replaced
	}
	return replaced
}
