package trace

import (
	"context"
	"sync"
	"time"

	"github.com/goshgame/gcomponent/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var (
	globalTracerProvider *TracerProvider
	once                 sync.Once
	defaultTestEndPoint  string = "a2a50f438d5374cc482d82269c04ee7a-758565007.ap-southeast-1.elb.amazonaws.com:4317"
	defaultProdEndPoint  string = "ab16de23a5e2f465f905241e7590d82c-431378024.ap-southeast-1.elb.amazonaws.com:4317"
)

// TracerProvider tracer provider
type TracerProvider struct {
	opts     *Options
	provider *sdktrace.TracerProvider
}

// NewTracerProvider get TracerProvider
func NewTracerProvider(opt ...Option) *TracerProvider {
	once.Do(func() {
		endPoint := defaultTestEndPoint
		if env.IsProd() {
			endPoint = defaultProdEndPoint
		}
		tp := &TracerProvider{}
		tp.opts = &Options{
			Endpoint:        endPoint,
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

// GetTracerProvider get TracerProvider
func GetTracerProvider() *TracerProvider {
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

// NewSpan new span
func (tp *TracerProvider) NewSpan(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	tr := tp.Tracer().Tracer("new-span")
	spanCtx, span := tr.Start(ctx, spanName, opts...)
	return spanCtx, span
}

// RunSpan run span
func (tp *TracerProvider) RunSpan(ctx context.Context, name string, fn func(ctx context.Context) error, attrs ...attribute.KeyValue) error {
	tr := tp.Tracer().Tracer("run-span")
	ctx, span := tr.Start(ctx, name)
	defer span.End()

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	start := time.Now()
	err := fn(ctx)
	span.SetAttributes(attribute.Int64("latency_ms", time.Since(start).Milliseconds()))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Int("biz_code", 1))
	} else {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(attribute.Int("biz_code", 0))
	}
	return err
}
