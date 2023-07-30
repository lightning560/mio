package gin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"miopkg/log"
	"miopkg/metric"
	"miopkg/trace"

	"go.opentelemetry.io/otel/propagation"
	otrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// extractAPP 提取header头中的app信息
func extractAID(ctx *gin.Context) string {
	return ctx.Request.Header.Get("AID")
}

func recoverMiddleware(logger *log.Logger, slowQueryThresholdInMilli int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		var beg = time.Now()
		var fields = make([]log.Field, 0, 8)
		var brokenPipe bool
		defer func() {
			//Latency
			fields = append(fields, zap.Float64("cost", time.Since(beg).Seconds()))
			if slowQueryThresholdInMilli > 0 {
				if cost := int64(time.Since(beg)) / 1e6; cost > slowQueryThresholdInMilli {
					fields = append(fields, zap.Int64("slow", cost))
				}
			}
			if rec := recover(); rec != nil {
				if ne, ok := rec.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				var err = rec.(error)
				fields = append(fields, zap.ByteString("stack", stack(3)))
				fields = append(fields, zap.String("err", err.Error()))
				logger.Error("access", fields...)
				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					c.Error(err) // nolint: errcheck
					c.Abort()
					return
				}
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			// httpRequest, _ := httputil.DumpRequest(c.Request, false)
			// fields = append(fields, zap.ByteString("request", httpRequest))
			fields = append(fields,
				zap.String("method", c.Request.Method),
				zap.Int("code", c.Writer.Status()),
				zap.Int("size", c.Writer.Size()),
				zap.String("host", c.Request.Host),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("err", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			)
			logger.Info("access", fields...)
		}()
		c.Next()
	}
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

// func timeFormat(t time.Time) string {
// 	timeString := t.Format("2006/01/02 - 15:04:05")
// 	return timeString
// }

func metricServerInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		beg := time.Now()
		c.Next()
		metric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeHTTP, c.Request.Method+"."+c.Request.URL.Path, extractAID(c))
		metric.ServerHandleCounter.Inc(metric.TypeHTTP, c.Request.Method+"."+c.Request.URL.Path, extractAID(c), http.StatusText(c.Writer.Status()))
	}
}

func traceServerInterceptor() gin.HandlerFunc {
	tracer := trace.NewTracer(otrace.SpanKindServer)
	return func(c *gin.Context) {
		// todo 该方法会在v0.9.0移除
		trace.CompatibleExtractHTTPTraceID(c.Request.Header)
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+"."+c.FullPath(), propagation.HeaderCarrier(c.Request.Header))
		span.SetAttributes(
			trace.TagComponent("http"),
			trace.TagSpanKind("server"),
			trace.CustomTag("http.url", c.Request.URL.Path),
			trace.CustomTag("http.target", c.FullPath()),
			trace.CustomTag("http.method", c.Request.Method),
			trace.CustomTag("net.peer.ip", c.ClientIP()),
		)
		c.Request = c.Request.WithContext(ctx)
		defer span.End()
		c.Header("application.TraceIDName()", span.SpanContext().TraceID().String())
		c.Next()
	}
}
