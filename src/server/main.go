package main

import (
	"encoding/json"
	"log"
	"net/http"
	"traefik-tryout/src/server/pkg/handler"
	"traefik-tryout/src/server/pkg/models"
	"traefik-tryout/src/server/pkg/repository"
	"traefik-tryout/src/server/pkg/service"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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
	defer closer.Close()

	span := tracer.StartSpan("server")

	r := repository.NewRepository(span)
	s := service.NewService(span, r)
	h := handler.NewHandler(span, s)

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		// Extract the context from the headers
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		serverSpan := tracer.StartSpan("server", ext.RPCServerOption(spanCtx))
		defer serverSpan.Finish()
		w.Write([]byte("siema"))
		w.WriteHeader(200)
	})

	http.HandleFunc("/customers", func(writer http.ResponseWriter, request *http.Request) {
		// Extract the context from the headers
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))
		serverSpan := tracer.StartSpan("server", ext.RPCServerOption(spanCtx))

		serverSpan.SetTag("tag", "testowy")
		defer serverSpan.Finish()
		switch request.Method {
		case http.MethodGet:
			json.NewEncoder(writer).Encode(h.WithTracer(serverSpan).GetCustomers())
		case http.MethodPost:
			var customer models.Customer
			if err := json.NewDecoder(request.Body).Decode(&customer); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
			} else {
				h.WithTracer(serverSpan).CreateCustomer(customer)
			}
		}
	})

	log.Fatal(http.ListenAndServe(":8082", nil))
}
