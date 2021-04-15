package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"traefik-tryout/src/server/pkg/models"

	"github.com/opentracing/opentracing-go/log"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func main() {
	// Sample configuration for testing. Use constant sampling to sample every trace
	// and enable LogSpan to log every span via configured Logger.
	cfg := jaegercfg.Configuration{
		ServiceName: "service",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		panic(err)
	}
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)

	defer closer.Close()
	// continue main()

	clientSpan := tracer.StartSpan("client")

	clientSpan.LogFields(
		log.String("service", "client"),
		log.String("log", "value-log"),
	)

	defer clientSpan.Finish()

	//url := "http://localhost:8082/publish"
	//req, _ := http.NewRequest("GET", url, nil)
	//
	//clientSpan.SetTag("tag", "test-tag")
	//clientSpan.LogKV("event", "setTag")
	//// Set some tags on the clientSpan to annotate that it's the client span.
	//// The additional HTTP tags are useful for debugging purposes.
	//ext.SpanKindRPCClient.Set(clientSpan)
	//ext.HTTPUrl.Set(clientSpan, url)
	//ext.HTTPMethod.Set(clientSpan, "GET")
	//
	//// Inject the client span context into the headers
	//if err := tracer.Inject(
	//	clientSpan.Context(),
	//	opentracing.HTTPHeaders,
	//	opentracing.HTTPHeadersCarrier(req.Header),
	//); err != nil {
	//	panic(err)
	//}
	//clientSpan.LogKV("event", "inject")
	//resp, _ := http.DefaultClient.Do(req)
	//fmt.Println(resp.StatusCode)

	s := ServerCLI{
		path: "http://localhost:8082/customers",
		h:    http.DefaultClient,
		t:    clientSpan,
	}
	s.CreateCustomer(models.Customer{
		ID:   1,
		Name: "Customer1",
	})
	fmt.Println(s.GetCustomers())
}

type ServerCLI struct {
	h    *http.Client
	path string
	t    opentracing.Span
}

func (s ServerCLI) CreateCustomer(customer models.Customer) {
	// Set some tags on the clientSpan to annotate that it's the client span.
	// The additional HTTP tags are useful for debugging purposes.
	ext.SpanKindRPCClient.Set(s.t)
	ext.HTTPUrl.Set(s.t, s.path)
	ext.HTTPMethod.Set(s.t, http.MethodPost)

	b, _ := json.Marshal(customer)
	req, err := http.NewRequest(http.MethodPost, s.path, bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf("failed, %s", err)
	}
	// Inject the client span context into the headers
	if err := s.t.Tracer().Inject(
		s.t.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	); err != nil {
		panic(err)
	}
	s.t.Finish()

	s.h.Do(req)
}

func (s ServerCLI) GetCustomers() []models.Customer {
	req, err := http.NewRequest(http.MethodGet, s.path, nil)
	if err != nil {
		fmt.Printf("failed, %s", err)
	}
	resp, err := s.h.Do(req)
	if err != nil {
		panic(err)
	}
	var c []models.Customer
	json.NewDecoder(resp.Body).Decode(&c)
	return c
}
