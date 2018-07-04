package tracing

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	cli "gopkg.in/urfave/cli.v1"
)

var Enabled bool = false

// TracingEnabledFlag is the CLI flag name to use to enable trace collections.
const TracingEnabledFlag = "tracing"

var (
	Closer io.Closer
)

var (
	TracingFlag = cli.BoolFlag{
		Name:  TracingEnabledFlag,
		Usage: "Enable tracing",
	}
	TracingEndpointFlag = cli.StringFlag{
		Name:  "tracing.endpoint",
		Usage: "Tracing endpoint",
		Value: "0.0.0.0:6831",
	}
)

// Flags holds all command-line flags required for tracing collection.
var Flags = []cli.Flag{
	TracingFlag,
	TracingEndpointFlag,
}

// Init enables or disables the metrics system. Since we need this to run before
// any other code gets to create meters and timers, we'll actually do an ugly hack
// and peek into the command line args for the metrics flag.
func init() {
	for _, arg := range os.Args {
		if flag := strings.TrimLeft(arg, "-"); flag == TracingEnabledFlag {
			Enabled = true
		}
	}
}

func Setup(ctx *cli.Context) {
	if Enabled {
		log.Info("Enabling opentracing")
		var (
			endpoint = ctx.GlobalString(TracingEndpointFlag.Name)
		)

		Closer = initTracer(endpoint)
	}
}

func initTracer(endpoint string) (closer io.Closer) {
	// Sample configuration for testing. Use constant sampling to sample every trace
	// and enable LogSpan to log every span via configured Logger.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  endpoint,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	//jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		"swarm",
		jaegercfg.Logger(jLogger),
		//jaegercfg.Metrics(jMetricsFactory),
		//jaegercfg.Observer(rpcmetrics.NewObserver(jMetricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		panic(fmt.Sprintf("Could not initialize jaeger tracer: %s", err.Error()))
	}

	return closer
}