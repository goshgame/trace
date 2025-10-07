package trace

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
)

var (
	globalTracerProvider *TracerProvider
	once                 sync.Once
)

type TracerProvider struct {
	opts     *Options
	provider *sdktrace.TracerProvider
}

// NewTracerProvider get TracerProvider
func NewTracerProvider(opt ...Option) *TracerProvider {
	once.Do(func() {
		tp := &TracerProvider{}
		tp.opts = &Options{
			Endpoint:        "localhost:43170",
			ServiceName:     "localtrace",
			BatchSize:       512,
			QueueSize:       8192,
			BatchTimeout:    1 * time.Second,
			ConnTimeout:     3 * time.Second,
			ShutdownTimeout: 3 * time.Second,
		}
		for _, o := range opt {
			o(tp.opts)
		}

		ctx, cancel := context.WithTimeout(context.Background(), tp.opts.ConnTimeout)
		defer cancel()
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(tp.opts.Endpoint),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithDialOption(
				grpc.WithBlock(),
			),
		)
		if err != nil {
			panic(err)
		}
		tp.provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()), // 全量采样
			sdktrace.WithBatcher(exporter,
				sdktrace.WithMaxExportBatchSize(tp.opts.BatchSize),
				sdktrace.WithMaxQueueSize(tp.opts.QueueSize),
				sdktrace.WithBatchTimeout(tp.opts.BatchTimeout),
			),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(tp.opts.ServiceName),
			)),
		)
		otel.SetTracerProvider(tp.provider)
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		)
		globalTracerProvider = tp
	})

	return globalTracerProvider
}

// Tracer get Tracer
func (tp *TracerProvider) Tracer() *sdktrace.TracerProvider {
	return tp.provider
}

// Close close
func (tp *TracerProvider) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), tp.opts.ShutdownTimeout)
	defer cancel()
	tp.provider.Shutdown(ctx)
}
