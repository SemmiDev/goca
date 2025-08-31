package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ulule/limiter/v3"

	. "github.com/smartystreets/goconvey/convey"
)

type mockRateLimiter struct {
	limit     int64
	period    time.Duration
	exceeded  bool
	remaining int64
	err       error
}

func (m *mockRateLimiter) Check(ctx context.Context, key string) (limiter.Context, error) {
	return limiter.Context{
		Limit:     m.limit,
		Remaining: m.remaining,
		Reached:   m.exceeded,
	}, m.err
}

func (m *mockRateLimiter) Take(ctx context.Context, key string) (limiter.Context, error) {
	return limiter.Context{
		Limit:     m.limit,
		Remaining: m.remaining,
		Reached:   m.exceeded,
	}, m.err
}

func (m *mockRateLimiter) GetLimit() int64 {
	return m.limit
}

func (m *mockRateLimiter) GetPeriod() time.Duration {
	return m.period
}

func TestMockRateLimiter(t *testing.T) {
	Convey("Diberikan mock RateLimiter", t, func() {
		ctx := context.Background()
		rl := &mockRateLimiter{
			limit:     10,
			period:    time.Minute,
			exceeded:  false,
			remaining: 9,
			err:       nil,
		}

		Convey("GetLimit dan GetPeriod bekerja", func() {
			So(rl.GetLimit(), ShouldEqual, 10)
			So(rl.GetPeriod(), ShouldEqual, time.Minute)
		})

		Convey("Check mengembalikan tidak terlampaui", func() {
			res, err := rl.Check(ctx, "user:1")
			So(err, ShouldBeNil)
			So(res.Reached, ShouldBeFalse)
			So(res.Remaining, ShouldEqual, 9)
		})

		Convey("Take mengembalikan error", func() {
			rl.err = errors.New("redis error")

			_, err := rl.Take(ctx, "user:1")
			So(err, ShouldNotBeNil)
		})

		Convey("Check mengembalikan status terlampaui", func() {
			rl.exceeded = true
			rl.remaining = 0

			res, err := rl.Check(ctx, "user:1")
			So(err, ShouldBeNil)
			So(res.Reached, ShouldBeTrue)
			So(res.Remaining, ShouldEqual, 0)
		})
	})
}
