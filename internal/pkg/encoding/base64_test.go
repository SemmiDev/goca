package encoding

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBase64Encoding(t *testing.T) {
	Convey("Diberikan fungsi-fungsi encoding base64", t, func() {
		originalData := []byte("hello world 123 !@#")
		encodedData := "aGVsbG8gd29ybGQgMTIzICFAIw=="

		Convey("Fungsi Base64Encode", func() {
			Convey("Ketika data byte yang valid diberikan", func() {
				result := Base64Encode(originalData)
				Convey("Seharusnya mengembalikan string base64 yang benar", func() {
					So(result, ShouldEqual, encodedData)
				})
			})
		})

		Convey("Fungsi Base64Decode", func() {
			Convey("Ketika string base64 yang valid diberikan", func() {
				decoded, err := Base64Decode(encodedData)

				Convey("Seharusnya tidak mengembalikan error", func() {
					So(err, ShouldBeNil)
				})
				Convey("Seharusnya mengembalikan data byte asli yang benar", func() {
					So(decoded, ShouldResemble, originalData)
				})
			})

			Convey("Ketika string base64 yang tidak valid diberikan", func() {
				_, err := Base64Decode("ini bukan base64 string###")

				Convey("Seharusnya mengembalikan error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}
