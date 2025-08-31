package random

import (
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateNumericOTP(t *testing.T) {
	Convey("Diberikan fungsi GenerateNumericOTP", t, func() {
		Convey("Ketika panjang yang valid diberikan", func() {
			length := 6
			otp, err := GenerateNumericOTP(length)

			Convey("Seharusnya tidak mengembalikan error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Seharusnya mengembalikan OTP dengan panjang yang benar", func() {
				So(len(otp), ShouldEqual, length)
			})

			Convey("Seharusnya mengembalikan OTP yang hanya berisi digit", func() {
				_, convErr := strconv.Atoi(otp)
				So(convErr, ShouldBeNil)
			})
		})

		Convey("Ketika panjangnya terlalu pendek", func() {
			length := minOTPLength - 1
			otp, err := GenerateNumericOTP(length)

			Convey("Seharusnya mengembalikan sebuah error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "OTP length must be between 4 and 12 digits")
			})

			Convey("Seharusnya mengembalikan string kosong", func() {
				So(otp, ShouldBeEmpty)
			})
		})

		Convey("Ketika panjangnya terlalu panjang", func() {
			length := maxOTPLength + 1
			otp, err := GenerateNumericOTP(length)

			Convey("Seharusnya mengembalikan sebuah error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "OTP length must be between 4 and 12 digits")
			})

			Convey("Seharusnya mengembalikan string kosong", func() {
				So(otp, ShouldBeEmpty)
			})
		})

		Convey("Ketika panjangnya adalah minimum yang diizinkan", func() {
			length := minOTPLength
			otp, err := GenerateNumericOTP(length)

			Convey("Seharusnya berhasil", func() {
				So(err, ShouldBeNil)
				So(len(otp), ShouldEqual, length)
			})
		})

		Convey("Ketika panjangnya adalah maksimum yang diizinkan", func() {
			length := maxOTPLength
			otp, err := GenerateNumericOTP(length)

			Convey("Seharusnya berhasil", func() {
				So(err, ShouldBeNil)
				So(len(otp), ShouldEqual, length)
			})
		})
	})
}
