package trace

import "time"

// Options - trace options
type Options struct {
	Endpoint        string
	ServiceName     string
	BatchSize       int
	QueueSize       int
	BatchTimeout    time.Duration
	ConnTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Option - trace option
type Option func(*Options)

// WithEndpoint 设置 endpoint
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

// WithServiceName 设置 serviceName
func WithServiceName(serviceName string) Option {
	return func(o *Options) {
		o.ServiceName = serviceName
	}
}

// WithBatchSize 设置 batchSize
func WithBatchSize(batchSize int) Option {
	return func(o *Options) {
		o.BatchSize = batchSize
	}
}

// WithQueueSize 设置 queueSize
func WithQueueSize(queueSize int) Option {
	return func(o *Options) {
		o.QueueSize = queueSize
	}
}

// WithBatchTimeout 设置 batchTimeout
func WithBatchTimeout(batchTimeout time.Duration) Option {
	return func(o *Options) {
		o.BatchTimeout = batchTimeout
	}
}

// WithConnTimeout 设置 connTimeout
func WithConnTimeout(connTimeout time.Duration) Option {
	return func(o *Options) {
		o.ConnTimeout = connTimeout
	}
}

// WithShutdownTimeout 设置 shutdownTimeout
func WithShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(o *Options) {
		o.ShutdownTimeout = shutdownTimeout
	}
}
