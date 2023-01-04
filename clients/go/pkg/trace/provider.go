// Copyright 2021 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package trace manages open telemetry trace exporter.
package trace

import (
	"context"
	"fmt"
	"time"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init initializes and sets the global trace exporter.
func Init(traceRatio float64) error {
	exporter, err := texporter.New()
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// According to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#parentbased
	// If the parent is sampled, it will be sampled here.
	// If the parent is not sampled, it won't be sampled here.
	// If this is the root span, it will be sampled with the given ratio below.
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(traceRatio))
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		// For google cloud specifically.
		propagator.CloudTraceFormatPropagator{},
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return nil
}

// Shutdown shuts down the global trace exporter.
func Shutdown() error {
	tpRaw := otel.GetTracerProvider()
	tp, ok := tpRaw.(*sdktrace.TracerProvider)
	if !ok {
		// Not the trace provider we expect, ignore.
		return nil
	}
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	if err := tp.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown trace provider: %w", err)
	}
	return nil
}
