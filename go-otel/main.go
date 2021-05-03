package main

import (
	"context"
	"log"
	"os"
	// "time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	// "go.opentelemetry.io/otel/trace"

	"fmt"
	"net/http"

	"google.golang.org/grpc/credentials"
)

// TODO: This is hacky
var provider *sdktrace.TracerProvider

func initTracer() func() {
	// Fetch the necessary settings (from environment variables, in this example).
	// You can find the API key via https://ui.honeycomb.io/account after signing up for Honeycomb.
	apikey, _ := os.LookupEnv("HONEYCOMB_API_KEY")
	dataset, _ := os.LookupEnv("HONEYCOMB_DATASET")

	// Initialize an OTLP exporter over gRPC and point it to Honeycomb.
	ctx := context.Background()
	exporter, err := otlp.NewExporter(
		ctx,
		otlpgrpc.NewDriver(
			otlpgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
			otlpgrpc.WithEndpoint("api.honeycomb.io:443"),
			otlpgrpc.WithHeaders(map[string]string{
				"x-honeycomb-team":    apikey,
				"x-honeycomb-dataset": dataset,
			}),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Configure the OTel tracer provider.
	provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(provider)

	// This callback will ensure all spans get flushed before the program exits.
	return func() {
		ctx := context.Background()
		err := provider.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tracer := provider.Tracer("mads-hartmann/o11y/handler")
	ctx, span := tracer.Start(ctx, "handler")

	span.SetAttributes(attribute.Any("id", "some-id"), attribute.Any("price", "some-price"))
	defer span.End()
	log.Printf("Handling request %s", r.URL.Path[1:])
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {

	cleanup := initTracer()

	ctx := context.Background()

	tracer := provider.Tracer("mads-hartmann/o11y/main")
	ctx, span := tracer.Start(ctx, "main")
	span.SetAttributes(attribute.Any("id", "some-id"), attribute.Any("price", "some-price"))
	defer span.End()

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	defer cleanup()
}
