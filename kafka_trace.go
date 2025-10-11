package trace

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectTraceToKafkaMessage - inject to kafka message
func InjectTraceToKafkaMessage(ctx context.Context, msg *kafka.Message) {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		msg.Headers = append(msg.Headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}
}

// ExtractTraceFromKafkaMessage - extract from kafka message
func ExtractTraceFromKafkaMessage(ctx context.Context, msg *kafka.Message) context.Context {
	carrier := propagation.MapCarrier{}
	for _, h := range msg.Headers {
		carrier[h.Key] = string(h.Value)
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
