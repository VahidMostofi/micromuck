package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/uber/jaeger-lib/metrics"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

const XXXL_PRIME = 1000000007 //6000-7000
const XL_PRIME = 45269999     //300-400
const L_PRIME = 9171787       //90-100
const M_PRIME = 3951371       //30-40
const S_PRIME = 999983

func isPrime(value uint) {
	var i uint
	for i = 2; i <= uint(math.Floor(float64(value)/2)); i++ {
		if value%i == 0 {
			return
		}
	}
	return
}

var (
	Service1 string
	Service2 string
	Name     string
	Port     string
	tracer   opentracing.Tracer
	closer   io.Closer
)

func initTracing() {
	cfg := jaegercfg.Configuration{
		ServiceName: Name,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: "jaeger:6831",
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	var err error
	tracer, closer, err = cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		panic(err)
	}
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)
}

func main() {
	Port = os.Getenv("PORT")
	Service1 = os.Getenv("SERVICE1")
	Service2 = os.Getenv("SERVICE2")
	Name = os.Getenv("NAME")

	initTracing()
	defer closer.Close()
	http.HandleFunc("/", SimpleServer)
	log("initialzed with", Port, Service1, Service2)
	http.ListenAndServe(":"+Port, nil)
	log("listening to port", Port)

}

func makeHTTPGET(tracer opentracing.Tracer, parent opentracing.SpanContext, name, url, debug_id string) (*http.Response, error) {
	clientSpan := tracer.StartSpan(name, ext.RPCServerOption(parent))
	defer clientSpan.Finish()

	req, _ := http.NewRequest("GET", url, nil)

	ext.SpanKindRPCClient.Set(clientSpan)
	ext.HTTPUrl.Set(clientSpan, url)
	ext.HTTPMethod.Set(clientSpan, "GET")

	req.Header.Add("debug_id", debug_id)

	tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	resp, err := http.DefaultClient.Do(req)
	return resp, err
}

func log(args ...interface{}) {
	fmt.Fprint(os.Stdout, Name, ": ", fmt.Sprintln(args...))
}

// SimpleServer for
func SimpleServer(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if r.URL.Path == "/gateway/service1" {
			tracer := opentracing.GlobalTracer()
			span := tracer.StartSpan("gateway")
			log("gateway service1")
			isPrime(S_PRIME)
			resp, err := makeHTTPGET(tracer, span.Context(), "GET_SERVICE1", Service1+"/service1", r.Header.Get("debug_id"))
			if err != nil || resp.StatusCode != 200 {
				log("ERROR", err)
				w.WriteHeader(500)
				span.Finish()
				return
			}
			isPrime(S_PRIME)
			w.Header().Add("debug_id", r.Header.Get("debug_id"))
			w.Write([]byte{})
			span.Finish()
			return
		} else if r.URL.Path == "/gateway/service2" {
			tracer := opentracing.GlobalTracer()
			span := tracer.StartSpan("gateway")
			log("gateway service2")
			isPrime(S_PRIME)
			resp, err := makeHTTPGET(tracer, span.Context(), "GET_SERVICE2", Service2+"/service2", r.Header.Get("debug_id"))
			if err != nil || resp.StatusCode != 200 {
				log("ERROR", err)
				w.WriteHeader(500)
				span.Finish()
				return
			}
			isPrime(S_PRIME)
			w.Header().Add("debug_id", r.Header.Get("debug_id"))
			w.Write([]byte{})
			span.Finish()
			return
		} else if r.URL.Path == "/service1" || r.URL.Path == "/service2" {
			// tracer := opentracing.GlobalTracer()
			spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
			serverSpan := tracer.StartSpan("server", ext.RPCServerOption(spanCtx))
			defer serverSpan.Finish()
			log(Name)
			if r.URL.Path == "/service2" {
				isPrime(S_PRIME)
			} else {
				isPrime(S_PRIME)
			}
			w.Header().Add("debug_id", r.Header.Get("debug_id"))
			w.Write([]byte{})
			return
		}
		w.WriteHeader(400)
	}
	w.WriteHeader(400)
}
