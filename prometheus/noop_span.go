package prometheus

import "github.com/arcgolabs/observabilityx"

type noopSpan struct{}

func (noopSpan) End() {}

func (noopSpan) RecordError(err error) {
	_ = err
}

func (noopSpan) SetAttributes(attrs ...observabilityx.Attribute) {
	_ = attrs
}
