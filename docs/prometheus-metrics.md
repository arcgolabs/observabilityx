---
title: 'observabilityx Prometheus metrics endpoint'
linkTitle: 'prometheus-metrics'
description: 'Expose /metrics using the Prometheus backend'
weight: 3
---

## Prometheus metrics endpoint

The Prometheus backend provides an HTTP handler via `promobs.Adapter.Handler()`. You can mount it on any router/framework.

Below is a minimal example mounting `/metrics` using `httpx` + `std` adapter (chi router).

## Example

```go
package main

import (
	"fmt"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	promobs "github.com/DaiYuANg/arcgo/observabilityx/prometheus"
)

func main() {
	prom := promobs.New(promobs.WithNamespace("app"))

	stdAdapter := std.New(nil, adapter.HumaOptions{DisableDocsRoutes: true})
	metricsServer := httpx.New(httpx.WithAdapter(stdAdapter))

	stdAdapter.Router().Handle("/metrics", prom.Handler())

	fmt.Println("metrics route registered: GET /metrics")
	_ = metricsServer
}
```

## Runnable example (repository)

- [examples/observabilityx/multi](https://github.com/DaiYuANg/arcgo/tree/main/examples/observabilityx/multi)
