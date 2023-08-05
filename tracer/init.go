package tracer

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/olaola-chat/rbp-library/tool"
	"github.com/olaola-chat/rbp-library/tracer/wrap"

	"github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/frame/g"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
)

var enpoitURL string

func init() {
	if g.Cfg().GetString("server.RunMode") == "dev" {
		//enpoitURL = "http://tracing-analysis-dc-hz.aliyuncs.com/adapt_ik4j6rki2p@87bb7ef3b9d9545_ik4j6rki2p@53df7ad2afe8301/api/traces"
		enpoitURL = "http://192.168.11.46:14268/api/traces?format=jaeger.thrift"
	} else {
		enpoitURL = "http://10.0.72.144:6834/api/traces?format=jaeger.thrift"
	}
	ip, err := tool.IP.LocalIPv4s()
	if err != nil {
		panic(err)
	}
	netIp := net.ParseIP(ip)
	if netIp == nil {
		panic(fmt.Errorf("error ip get %s", ip))
	}
	name := g.Cfg().GetString("server.TraceName")
	if len(name) == 0 {
		name = "test"
	}
	opentracing.InitGlobalTracer(getJaegerTracer(name, binary.BigEndian.Uint32(netIp.To4())))

	//msyql 注入
	sql.Register(
		"bbsql",
		wrap.Driver(
			mysql.MySQLDriver{},
		),
	)
}

func getJaegerTracer(serviceName string, ip uint32) opentracing.Tracer {
	sender := transport.NewHTTPTransport(
		enpoitURL,
	)
	tracer, _ := jaeger.NewTracer(
		serviceName,
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(
			sender,
			jaeger.ReporterOptions.Logger(jaeger.StdLogger),
		),
	)
	return tracer
}

/*
func getTracer(serviceName string, ip string, samplerValue uint64) opentracing.Tracer {
	reporter := httpreporter.NewReporter(
		enpoitURL,
		httpreporter.BatchSize(500),
	)
	endpoint, _ := zipkin.NewEndpoint(serviceName, ip)
	sampler := zipkin.NewModuloSampler(samplerValue)
	nativeTracer, _ := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSampler(sampler),
	)
	return zipkinot.Wrap(nativeTracer)
}
*/
