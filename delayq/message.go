package delayq

import (
	"bytes"
	"context"
	"encoding/gob"

	ztrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type (
	Action  string
	Message struct {
		Carrier propagation.MapCarrier
		Action  Action
		Body    []byte
	}
)

func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *Message) Inject(ctx context.Context) (context.Context, trace.Span) {
	tracer := otel.Tracer(ztrace.TraceName)
	spanCtx, span := tracer.Start(ctx, string(m.Action), trace.WithSpanKind(trace.SpanKindProducer))
	m.Carrier = propagation.MapCarrier{}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(spanCtx, m.Carrier)
	return spanCtx, span
}

func (m *Message) Decode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(m)
}

func (m *Message) Extract(ctx context.Context) (context.Context, trace.Span) {
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, m.Carrier)
	tracer := otel.Tracer(ztrace.TraceName)
	spanCtx, span := tracer.Start(ctx, string(m.Action), trace.WithSpanKind(trace.SpanKindConsumer))
	return spanCtx, span
}
