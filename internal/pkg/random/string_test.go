package random

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRandomStringGeneration(t *testing.T) {

	Convey("Diberikan fungsi GenerateString", t, func() {
		Convey("Ketika panjang dan charset yang valid diberikan", func() {
			length := 16
			charset := CharsetAlphaNumeric
			result, err := GenerateString(length, charset)

			Convey("Seharusnya tidak mengembalikan error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Hasilnya harus memiliki panjang yang benar", func() {
				So(len(result), ShouldEqual, length)
			})
			Convey("Semua karakter dalam hasil harus berasal dari charset yang diberikan", func() {
				for _, char := range result {
					So(strings.ContainsRune(charset, char), ShouldBeTrue)
				}
			})
		})

		Convey("Ketika panjangnya nol", func() {
			result, err := GenerateString(0, CharsetAlphaNumeric)
			Convey("Seharusnya mengembalikan sebuah error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "length must be greater than zero")
			})
			Convey("Hasilnya harus berupa string kosong", func() {
				So(result, ShouldBeEmpty)
			})
		})

		Convey("Ketika charset kosong", func() {
			result, err := GenerateString(10, "")
			Convey("Seharusnya mengembalikan sebuah error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "charset must not be empty")
			})
			Convey("Hasilnya harus berupa string kosong", func() {
				So(result, ShouldBeEmpty)
			})
		})
	})

	Convey("Diberikan fungsi pembantu String", t, func() {
		Convey("Ketika dipanggil dengan panjang yang valid", func() {
			length := 20
			result, err := String(length)

			Convey("Seharusnya menghasilkan string alfanumerik dengan sukses", func() {
				So(err, ShouldBeNil)
				So(len(result), ShouldEqual, length)
				for _, char := range result {
					So(strings.ContainsRune(CharsetAlphaNumeric, char), ShouldBeTrue)
				}
			})
		})
	})

	Convey("Diberikan fungsi pembantu HexToken", t, func() {
		Convey("Ketika dipanggil dengan panjang yang valid", func() {
			length := 32
			result, err := HexToken(length)

			Convey("Seharusnya menghasilkan string hex dengan sukses", func() {
				So(err, ShouldBeNil)
				So(len(result), ShouldEqual, length)
				for _, char := range result {
					So(strings.ContainsRune(CharsetHex, char), ShouldBeTrue)
				}
			})
		})
	})
}
