---
title: 'observabilityx Prometheus metrics endpoint'
linkTitle: 'prometheus-metrics'
description: 'Expose /metrics using the Prometheus backend'
weight: 3
---

## Prometheus metrics endpoint

The Prometheus backend provides an HTTP handler via `promobs.Adapter.Handler()`. You can mount it on any router/framework.

Below is a minimal example mounting `/metrics` using the standard library HTTP mux.

## Example

```go
package main

import (
	"fmt"
	"net/http"

	promobs "github.com/arcgolabs/observabilityx/prometheus"
)

func main() {
	prom := promobs.New(promobs.WithNamespace("app"))

	http.Handle("/metrics", prom.Handler())

	fmt.Println("metrics route registered: GET /metrics")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
```
