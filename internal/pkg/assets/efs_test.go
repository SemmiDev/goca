package assets

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEmbeddedFiles(t *testing.T) {
	Convey("Given the embedded file system", t, func() {
		Convey("When checking if EmbeddedFiles is properly initialized", func() {
			Convey("Then it should not be nil", func() {
				So(EmbeddedFiles, ShouldNotBeNil)
			})
		})

		Convey("When checking for email verification template", func() {
			Convey("Then the file should exist and be readable", func() {
				data, err := EmbeddedFiles.ReadFile(EmailVerificationTemplatePath)
				So(err, ShouldBeNil)
				So(len(data), ShouldBeGreaterThan, 0)
			})
		})

		Convey("When checking for forgot password template", func() {
			Convey("Then the file should exist and be readable", func() {
				data, err := EmbeddedFiles.ReadFile(EmailForgotPasswordTemplatePath)
				So(err, ShouldBeNil)
				So(len(data), ShouldBeGreaterThan, 0)
			})
		})

		Convey("When checking for non-existent file", func() {
			Convey("Then it should return an error", func() {
				_, err := EmbeddedFiles.ReadFile("emails/non-existent.tmpl")
				So(err, ShouldNotBeNil)
			})
		})
	})
}
