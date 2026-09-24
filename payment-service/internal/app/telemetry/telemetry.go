package telemetry

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var service string
var logger *log.Logger
var requests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "demo_http_requests_total", Help: "HTTP requests by normalized route"}, []string{"service", "method", "route", "status"})
var duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "demo_http_request_duration_seconds", Help: "HTTP request latency", Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2, 5}}, []string{"service", "method", "route"})
var Registry = prometheus.NewRegistry()

func Init(name string) func(context.Context) error {
	service = name
	dir := os.Getenv("LOG_DIR")
	if dir == "" {
		dir = "storage/logs"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(dir, name+".json"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(f, "", 0)
	exporter, err := otlptracehttp.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithResource(resource.NewWithAttributes("", attribute.String("service.name", name))))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	Registry.MustRegister(requests, duration)
	return func(ctx context.Context) error { defer f.Close(); return tp.Shutdown(ctx) }
}
func Event(ctx context.Context, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	sc := trace.SpanContextFromContext(ctx)
	fields["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	fields["service"] = service
	fields["message"] = message
	fields["trace_id"] = sc.TraceID().String()
	fields["span_id"] = sc.SpanID().String()
	b, _ := json.Marshal(fields)
	logger.Print(string(b))
}
func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/metrics" {
			return c.Next()
		}
		start := time.Now()
		method := strings.Clone(c.Method())
		carrier := propagation.MapCarrier{}
		c.Request().Header.VisitAll(func(k, v []byte) { carrier.Set(string(k), string(v)) })
		// Fiber header casing is normalized; explicitly copy trace context.
		carrier.Set("traceparent", c.Get("traceparent"))
		carrier.Set("tracestate", c.Get("tracestate"))
		ctx := otel.GetTextMapPropagator().Extract(c.UserContext(), carrier)
		ctx, span := otel.Tracer(service).Start(ctx, method, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		c.SetUserContext(ctx)
		c.Set("X-Trace-ID", span.SpanContext().TraceID().String())
		err := c.Next()
		if err != nil {
			_ = c.App().ErrorHandler(c, err)
		}
		route := strings.Clone(c.Route().Path)
		if route == "" {
			route = "unmatched"
		}
		status := c.Response().StatusCode()
		span.SetName(method + " " + route)
		span.SetAttributes(attribute.String("http.request.method", method), attribute.String("http.route", route), attribute.Int("http.response.status_code", status))
		if status >= 400 {
			span.SetStatus(codes.Error, strconv.Itoa(status))
		}
		requests.WithLabelValues(service, method, route, strconv.Itoa(status)).Add(1)
		duration.WithLabelValues(service, method, route).Observe(time.Since(start).Seconds())
		Event(ctx, "http_request", map[string]interface{}{"method": c.Method(), "route": route, "status": status, "duration_ms": float64(time.Since(start).Microseconds()) / 1000})
		return nil
	}
}
func InstrumentDB(db *gorm.DB) {
	before := func(tx *gorm.DB) {
		ctx, span := otel.Tracer(service).Start(tx.Statement.Context, "postgresql", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("db.system", "postgresql")))
		tx.Statement.Context = ctx
		tx.InstanceSet("demo_span", span)
	}
	after := func(tx *gorm.DB) {
		if v, ok := tx.InstanceGet("demo_span"); ok {
			span := v.(trace.Span)
			span.SetAttributes(attribute.String("db.collection.name", tx.Statement.Table))
			if tx.Error != nil {
				span.RecordError(tx.Error)
				span.SetStatus(codes.Error, "database operation failed")
			}
			span.End()
		}
	}
	db.Callback().Create().Before("gorm:create").Register("demo:before", before)
	db.Callback().Create().After("gorm:create").Register("demo:after", after)
	db.Callback().Query().Before("gorm:query").Register("demo:before", before)
	db.Callback().Query().After("gorm:query").Register("demo:after", after)
	db.Callback().Update().Before("gorm:update").Register("demo:before", before)
	db.Callback().Update().After("gorm:update").Register("demo:after", after)
}
