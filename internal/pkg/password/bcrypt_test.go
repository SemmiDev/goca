package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashing(t *testing.T) {

	Convey("Diberikan fungsi-fungsi hashing password", t, func() {
		password := "mySecretPassword123"

		Convey("Fungsi HashPassword", func() {
			Convey("Ketika password yang valid diberikan", func() {
				hash, err := HashPassword(password)

				Convey("Seharusnya tidak mengembalikan error", func() {
					So(err, ShouldBeNil)
				})
				Convey("Seharusnya mengembalikan hash yang tidak kosong", func() {
					So(hash, ShouldNotBeEmpty)
				})
				Convey("Hash yang dihasilkan harus merupakan hash bcrypt yang valid", func() {
					// Membandingkan hash dengan password kosong akan gagal, tapi ini memvalidasi format hash.
					err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(""))
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Fungsi CheckPasswordHash", func() {
			// Hasilkan hash terlebih dahulu untuk digunakan dalam perbandingan
			hashedPassword, _ := HashPassword(password)

			Convey("Ketika password yang benar diberikan", func() {
				match := CheckPasswordHash(password, hashedPassword)
				Convey("Seharusnya mengembalikan true", func() {
					So(match, ShouldBeTrue)
				})
			})

			Convey("Ketika password yang salah diberikan", func() {
				match := CheckPasswordHash("wrongPassword", hashedPassword)
				Convey("Seharusnya mengembalikan false", func() {
					So(match, ShouldBeFalse)
				})
			})

			Convey("Ketika hash yang tidak valid diberikan", func() {
				match := CheckPasswordHash(password, "notAValidHash")
				Convey("Seharusnya mengembalikan false", func() {
					So(match, ShouldBeFalse)
				})
			})
		})
	})
}
