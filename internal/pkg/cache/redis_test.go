package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRedisClient(t *testing.T) {
	Convey("Diberikan Redis Cache Client", t, func() {
		db, mock := redismock.NewClientMock()
		redisClient := &RedisClient{Client: db}
		ctx := context.Background()

		key := "test_key"
		value := "test_value"
		exp := time.Minute

		Convey("Metode Set", func() {
			Convey("Seharusnya berhasil menyimpan nilai jika tidak ada error dari Redis", func() {
				mock.ExpectSet(key, value, exp).SetVal("OK")
				err := redisClient.Set(ctx, key, value, exp)
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("Seharusnya mengembalikan error jika Redis gagal menyimpan nilai", func() {
				mock.ExpectSet(key, value, exp).SetErr(errors.New("redis error"))
				err := redisClient.Set(ctx, key, value, exp)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "redis error")
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})
		})

		Convey("Metode Get", func() {
			Convey("Seharusnya berhasil mendapatkan nilai jika kunci ada", func() {
				mock.ExpectGet(key).SetVal(value)
				result, err := redisClient.Get(ctx, key)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, value)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("Seharusnya mengembalikan error redis.Nil jika kunci tidak ada", func() {
				mock.ExpectGet(key).RedisNil()
				_, err := redisClient.Get(ctx, key)
				So(err, ShouldEqual, redis.Nil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})
		})

		Convey("Metode Delete", func() {
			Convey("Seharusnya berhasil menghapus kunci jika ada", func() {
				mock.ExpectDel(key).SetVal(1) // 1 artinya 1 kunci dihapus
				err := redisClient.Delete(ctx, key)
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("Seharusnya mengembalikan error jika Redis gagal menghapus", func() {
				mock.ExpectDel(key).SetErr(errors.New("delete failed"))
				err := redisClient.Delete(ctx, key)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "delete failed")
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})
		})
	})
}
