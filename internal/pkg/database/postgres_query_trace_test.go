package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sammidev/goca/internal/pkg/logger"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestQueryTracer(t *testing.T) {
	Convey("Diberikan queryTracer dengan logger yang diamati", t, func() {
		core, recorded := observer.New(zapcore.DebugLevel)
		logger := &mockTracerLogger{
			SugaredLogger: zap.New(core).Sugar(),
		}

		tracer := &queryTracer{logger: logger}
		ctx := context.Background()

		Convey("TraceQueryStart", func() {
			sql := "SELECT * FROM users WHERE id = $1"
			args := []any{1}
			traceStartData := pgx.TraceQueryStartData{
				SQL:  sql,
				Args: args,
			}

			newCtx := tracer.TraceQueryStart(ctx, nil, traceStartData)
			So(newCtx, ShouldNotBeNil)

			td, ok := newCtx.Value(traceDataKey{}).(*traceData)
			So(ok, ShouldBeTrue)
			So(td, ShouldNotBeNil)
			So(td.SQL, ShouldEqual, sql)
			So(td.Args, ShouldResemble, args)
			So(td.queryType, ShouldEqual, "SELECT")
		})

		Convey("TraceQueryEnd", func() {
			sql := "INSERT INTO users (name) VALUES ($1)"
			args := []any{"test"}
			startTime := time.Now()

			baseCtx := context.WithValue(ctx, traceDataKey{}, &traceData{
				startTime: startTime,
				queryType: "INSERT",
				SQL:       sql,
				Args:      args,
			})

			Convey("ketika kueri berhasil", func() {
				cmdTag := pgconn.NewCommandTag("INSERT 0 1")
				traceEndData := pgx.TraceQueryEndData{
					CommandTag: cmdTag,
					Err:        nil,
				}
				tracer.TraceQueryEnd(baseCtx, nil, traceEndData)

				So(recorded.Len(), ShouldEqual, 1)
				log := recorded.All()[0]
				So(log.Level, ShouldEqual, zapcore.DebugLevel)
				So(log.Message, ShouldEqual, "Query executed successfully")
				So(log.ContextMap()["rows_affected"], ShouldEqual, int64(1))
				So(log.ContextMap()["query"], ShouldEqual, sql)
			})

			Convey("ketika kueri gagal", func() {
				err := errors.New("gagal terhubung")
				traceEndData := pgx.TraceQueryEndData{Err: err}
				tracer.TraceQueryEnd(baseCtx, nil, traceEndData)

				So(recorded.Len(), ShouldEqual, 1)
				log := recorded.All()[0]
				So(log.Level, ShouldEqual, zapcore.ErrorLevel)
				So(log.Message, ShouldEqual, "Query failed")
				So(log.ContextMap()["error"], ShouldEqual, err.Error())
			})

			Convey("ketika kueri tidak mengembalikan baris", func() {
				traceEndData := pgx.TraceQueryEndData{Err: pgx.ErrNoRows}
				tracer.TraceQueryEnd(baseCtx, nil, traceEndData)

				So(recorded.Len(), ShouldEqual, 1)
				log := recorded.All()[0]
				So(log.Level, ShouldEqual, zapcore.DebugLevel)
				So(log.Message, ShouldEqual, "Query returned no rows")
			})

			Convey("ketika data jejak tidak ada dalam konteks", func() {
				tracer.TraceQueryEnd(context.Background(), nil, pgx.TraceQueryEndData{})
				So(recorded.Len(), ShouldEqual, 1)
				log := recorded.All()[0]
				So(log.Level, ShouldEqual, zapcore.WarnLevel)
				So(log.Message, ShouldEqual, "Trace data not found in context; skipping query logging")
			})
		})

		Convey("extractQueryType", func() {
			So(extractQueryType("SELECT * FROM users"), ShouldEqual, "SELECT")
			So(extractQueryType("INSERT INTO users..."), ShouldEqual, "INSERT")
			So(extractQueryType("UPDATE users SET..."), ShouldEqual, "UPDATE")
			So(extractQueryType("DELETE FROM users..."), ShouldEqual, "DELETE")
			So(extractQueryType("CREATE TABLE users..."), ShouldEqual, "OTHER")
			So(extractQueryType(""), ShouldEqual, "UNKNOWN")
		})
	})
}

// mockTracerLogger adalah logger palsu untuk menguji pelacak kueri
type mockTracerLogger struct {
	*zap.SugaredLogger
}

func (m *mockTracerLogger) WithComponent(component string) logger.Logger {
	return &mockTracerLogger{m.SugaredLogger.With("component", component)}
}
func (m *mockTracerLogger) WithContext(ctx context.Context) logger.Logger { return m }
func (m *mockTracerLogger) With(keysAndValues ...interface{}) logger.Logger {
	return &mockTracerLogger{m.SugaredLogger.With(keysAndValues...)}
}

func (m *mockTracerLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.SugaredLogger.Fatalw(msg, keysAndValues...)
}

func (m *mockTracerLogger) Error(msg string, keysAndValues ...interface{}) {
	m.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (m *mockTracerLogger) Info(msg string, keysAndValues ...interface{}) {
	m.SugaredLogger.Infow(msg, keysAndValues...)
}

func (m *mockTracerLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.SugaredLogger.Warnw(msg, keysAndValues...)
}

func (m *mockTracerLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.SugaredLogger.Debugw(msg, keysAndValues...)
}
