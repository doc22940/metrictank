package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/grafana/metrictank/util"

	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	jaeger "github.com/uber/jaeger-client-go"
	macaron "gopkg.in/macaron.v1"
)

var (
	LogHeaders = false
)

type LoggingResponseWriter struct {
	macaron.ResponseWriter
	errBody *bytes.Buffer // the body in case it is an error
}

func (rw *LoggingResponseWriter) Write(b []byte) (int, error) {
	if rw.ResponseWriter.Status() >= 400 {
		rw.errBody.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

func Logger() macaron.Handler {
	return func(ctx *Context) {
		start := time.Now()
		rw := &LoggingResponseWriter{
			ResponseWriter: ctx.Resp,
			errBody:        &bytes.Buffer{},
		}
		ctx.Resp = rw
		ctx.MapTo(ctx.Resp, (*http.ResponseWriter)(nil))
		ctx.Next()

		// Only log:
		// - requests that resulted in errors
		// - requests on /render path
		if rw.Status() >= 200 && rw.Status() < 300 && !strings.HasPrefix(ctx.Req.URL.Path, "/render") {
			return
		}

		var content strings.Builder
		fmt.Fprintf(&content, "ts=%s", time.Now().Format(time.RFC3339Nano))

		traceID, sampled := extractTraceID(ctx.Req.Context())
		if traceID != "" {
			fmt.Fprintf(&content, " traceID=%s, sampled=%t", traceID, sampled)
		}

		err := ctx.Req.ParseForm()
		if err != nil {
			log.Errorf("Could not parse http request: %v", err)
		}
		paramsAsString := ""
		if len(ctx.Req.Form) > 0 {
			paramsAsString += "?"
			paramsAsString += ctx.Req.Form.Encode()
		}

		fmt.Fprintf(&content, " msg=\"%s %s%s (%v) %v\" orgID=%d", ctx.Req.Method, ctx.Req.URL.Path, paramsAsString, rw.Status(), time.Since(start), ctx.OrgId)

		referer := ctx.Req.Referer()
		if referer != "" {
			fmt.Fprintf(&content, " referer=%s", referer)
		}
		sourceIP := ctx.RemoteAddr()
		if sourceIP != "" {
			fmt.Fprintf(&content, " sourceIP=\"%s\"", sourceIP)
		}

		if rw.Status() < 200 || rw.Status() >= 300 {
			var errBody string
			if rw.Header().Get("Content-Encoding") == "gzip" {
				errBody, err = util.DecompressGzip(rw.errBody)
				if err != nil {
					log.Errorf("Decompressing gzip body failed: %s", err.Error())
				}
			} else {
				errBody = rw.errBody.String()
			}
			errorMsg := url.PathEscape(errBody)
			fmt.Fprintf(&content, " error=\"%s\"", errorMsg)
		}

		if LogHeaders {
			headers, err := extractHeaders(ctx.Req.Request)
			if err != nil {
				log.Errorf("Could not extract request headers: %v", err)
			}
			fmt.Fprintf(&content, " headers=\"%s\"", string(headers))
		}

		log.Println(content.String())
	}
}

func extractHeaders(req *http.Request) (string, error) {
	var b bytes.Buffer

	// Exclude some headers for security, or just that we don't need them when debugging
	err := req.Header.WriteSubset(&b, map[string]bool{
		"Cookie":        true,
		"X-Csrf-Token":  true,
		"Authorization": true,
	})
	if err != nil {
		return "", err
	}
	return url.PathEscape(string(b.Bytes())), nil
}

func extractTraceID(ctx context.Context) (string, bool) {
	sp := opentracing.SpanFromContext(ctx)
	if sp == nil {
		return "", false
	}
	sctx, ok := sp.Context().(jaeger.SpanContext)
	if !ok {
		return "", false
	}

	return sctx.TraceID().String(), sctx.IsSampled()
}
