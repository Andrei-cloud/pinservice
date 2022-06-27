package service

import (
	"context"
	"flag"
	"fmt"
	"net"
	http2 "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrei-cloud/pinservice/pkg/broker"
	endpoint "github.com/andrei-cloud/pinservice/pkg/endpoint"
	http1 "github.com/andrei-cloud/pinservice/pkg/http"
	"github.com/andrei-cloud/pinservice/pkg/pool"
	service "github.com/andrei-cloud/pinservice/pkg/service"
	endpoint1 "github.com/go-kit/kit/endpoint"
	prometheus "github.com/go-kit/kit/metrics/prometheus"
	log "github.com/go-kit/log"
	group "github.com/oklog/oklog/pkg/group"
	opentracinggo "github.com/opentracing/opentracing-go"
	zipkingoopentracing "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkingo "github.com/openzipkin/zipkin-go"
	http "github.com/openzipkin/zipkin-go/reporter/http"
	prometheus1 "github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

var tracer opentracinggo.Tracer
var logger log.Logger

// Define our flags. Your service probably won't need to bind listeners for
// all* supported transports, but we do it here for demonstration purposes.
var fs = flag.NewFlagSet("pin", flag.ExitOnError)
var hsmAddr = fs.String("hsm-addr", ":1500", "Thales HSM address")
var debugAddr = fs.String("debug-addr", ":8080", "Debug and metrics listen address")
var httpAddr = fs.String("http-addr", ":8081", "HTTP listen address")
var zipkinURL = fs.String("zipkin-url", "", "Enable Zipkin tracing via a collector URL e.g. http://localhost:9411/api/v1/spans")

func Run() {
	fs.Parse(os.Args[1:])

	// Create a single logger, which we'll use and give to other components.
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestamp)
	logger = log.With(logger, "caller", log.DefaultCaller)

	//  Determine which tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency
	if *zipkinURL != "" {
		logger.Log("tracer", "Zipkin", "URL", *zipkinURL)
		reporter := http.NewReporter(*zipkinURL)
		defer reporter.Close()
		endpoint, err := zipkingo.NewEndpoint("pin", "localhost:80")
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		localEndpoint := zipkingo.WithLocalEndpoint(endpoint)
		nativeTracer, err := zipkingo.NewTracer(reporter, localEndpoint)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		tracer = zipkingoopentracing.Wrap(nativeTracer)
	} else {
		logger.Log("tracer", "none")
		tracer = opentracinggo.GlobalTracer()
	}

	//define pool factory function
	factory := func(addr string) pool.Factory {
		return func() (pool.PoolItem, error) {
			return net.Dial("tcp", addr)
		}
	}

	logger.Log("pool", 2)
	p := pool.NewPool(2, factory(*hsmAddr))

	logger.Log("broker", 2)
	hsmBroker := broker.NewBroker(p, 2, logger)
	defer hsmBroker.Close()

	brokerCtx, stopBroker := context.WithCancel(context.Background())
	defer stopBroker()

	go hsmBroker.Start(brokerCtx)

	svc := service.New(hsmBroker, getServiceMiddleware(logger))
	eps := endpoint.New(svc, getEndpointMiddleware(logger))
	g := createService(eps)
	initMetricsEndpoint(g)
	initCancelInterrupt(g)
	logger.Log("exit", g.Run())

}
func initHttpHandler(endpoints endpoint.Endpoints, g *group.Group) {
	options := defaultHttpOptions(logger, tracer)
	// Add your http options here

	httpHandler := http1.NewHTTPHandler(endpoints, options)
	httpListener, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		return http2.Serve(httpListener, httpHandler)
	}, func(error) {
		httpListener.Close()
	})

}
func getServiceMiddleware(logger log.Logger) (mw []service.Middleware) {
	mw = []service.Middleware{}
	mw = addDefaultServiceMiddleware(logger, mw)
	// Append your middleware here

	return
}
func getEndpointMiddleware(logger log.Logger) (mw map[string][]endpoint1.Middleware) {
	mw = map[string][]endpoint1.Middleware{}
	duration := prometheus.NewSummaryFrom(prometheus1.SummaryOpts{
		Help:      "Request duration in seconds.",
		Name:      "request_duration_seconds",
		Namespace: "cards",
		Subsystem: "pin",
		Objectives: map[float64]float64{
			0:    0.1,
			0.25: 0.15,
			0.5:  0.1,
			0.75: 0.05,
			0.99: 0.001,
		},
	}, []string{"method", "success"})
	addDefaultEndpointMiddleware(logger, duration, mw)
	// Add you endpoint middleware here

	return
}
func initMetricsEndpoint(g *group.Group) {
	http2.DefaultServeMux.Handle("/metrics", promhttp.Handler())
	debugListener, err := net.Listen("tcp", *debugAddr)
	if err != nil {
		logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "debug/HTTP", "addr", *debugAddr)
		return http2.Serve(debugListener, http2.DefaultServeMux)
	}, func(error) {
		debugListener.Close()
	})
}
func initCancelInterrupt(g *group.Group) {
	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})
}
