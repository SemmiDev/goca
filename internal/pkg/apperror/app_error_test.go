package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/sammidev/goca/internal/pkg/validator"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAppErrors(t *testing.T) {

	Convey("Diberikan struktur dan fungsi-fungsi AppError", t, func() {

		Convey("Metode AppError.Error()", func() {
			Convey("Seharusnya memformat pesan dengan benar tanpa detail", func() {
				err := NewAppError(ErrCodeNotFound, "Resource not found")
				expected := "NOT_FOUND: Resource not found"
				So(err.Error(), ShouldEqual, expected)
			})

			Convey("Seharusnya memformat pesan dengan benar dengan detail", func() {
				err := NewAppErrorWithDetails(ErrCodeInvalidInput, "Invalid ID", "ID must be a UUID")
				expected := "INVALID_INPUT: Invalid ID (ID must be a UUID)"
				So(err.Error(), ShouldEqual, expected)
			})
		})

		Convey("Metode AppError.Unwrap()", func() {
			Convey("Seharusnya mengembalikan nil ketika tidak ada cause", func() {
				appErr := NewAppError(ErrCodeInternalError, "Something went wrong")
				So(appErr.Unwrap(), ShouldBeNil)
				So(errors.Unwrap(appErr), ShouldBeNil)
			})

			Convey("Seharusnya mengembalikan error yang dibungkus", func() {
				cause := errors.New("underlying database error")
				appErr := NewAppErrorWithCause(ErrCodeDatabaseError, "DB query failed", cause)
				So(appErr.Unwrap(), ShouldEqual, cause)
				So(errors.Unwrap(appErr), ShouldEqual, cause)
			})
		})

		Convey("Metode AppError.Is()", func() {
			err1 := NewAppError(ErrCodeNotFound, "message 1")
			err2 := NewAppError(ErrCodeNotFound, "message 2")
			err3 := NewAppError(ErrCodeInvalidInput, "message 3")

			Convey("Seharusnya mengembalikan true untuk error dengan kode yang sama", func() {
				So(errors.Is(err1, err2), ShouldBeTrue)
			})
			Convey("Seharusnya mengembalikan false untuk error dengan kode yang berbeda", func() {
				So(errors.Is(err1, err3), ShouldBeFalse)
			})
		})

		Convey("Metode AppError.HTTPStatusCode()", func() {
			testCases := []struct {
				name       string
				err        *AppError
				expectedSC int
			}{
				{"NotFound", NewAppError(ErrCodeNotFound, ""), http.StatusNotFound},
				{"InvalidInput", NewAppError(ErrCodeInvalidInput, ""), http.StatusBadRequest},
				{"Unauthorized", NewAppError(ErrCodeUnauthorized, ""), http.StatusUnauthorized},
				{"Forbidden", NewAppError(ErrCodeForbidden, ""), http.StatusForbidden},
				{"Conflict", NewAppError(ErrCodeConflict, ""), http.StatusConflict},
				{"InternalError", NewAppError(ErrCodeInternalError, ""), http.StatusInternalServerError},
				{"UserNotFound", NewAppError(ErrCodeUserNotFound, ""), http.StatusNotFound},
				{"DefaultInternalError", NewAppError(ErrorCode("UNKNOWN_CODE"), ""), http.StatusInternalServerError},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("Untuk error %s, seharusnya mengembalikan %d", tc.name, tc.expectedSC), func() {
					So(tc.err.HTTPStatusCode(), ShouldEqual, tc.expectedSC)
				})
			}
		})

		Convey("Fungsi-fungsi pembantu", func() {
			Convey("IsAppError", func() {
				appErr := NewAppError(ErrCodeInternalError, "test")
				stdErr := errors.New("standard error")

				Convey("Seharusnya mengembalikan true dan error ketika AppError diberikan", func() {
					extracted, ok := IsAppError(appErr)
					So(ok, ShouldBeTrue)
					So(extracted, ShouldEqual, appErr)
				})

				Convey("Seharusnya mengembalikan false dan nil ketika error standar diberikan", func() {
					extracted, ok := IsAppError(stdErr)
					So(ok, ShouldBeFalse)
					So(extracted, ShouldBeNil)
				})
			})

			Convey("IsValidationError", func() {
				validationErr := NewValidationError(validator.ValidationErrors{})
				appErr := NewAppError(ErrCodeInternalError, "test")

				Convey("Seharusnya mengembalikan true dan error ketika ValidationErrors diberikan", func() {
					extracted, ok := IsValidationError(validationErr)
					So(ok, ShouldBeTrue)
					So(extracted, ShouldEqual, validationErr)
				})
				Convey("Seharusnya mengembalikan false dan nil ketika AppError biasa diberikan", func() {
					extracted, ok := IsValidationError(appErr)
					So(ok, ShouldBeFalse)
					So(extracted, ShouldBeNil)
				})
			})

			Convey("WrapError", func() {
				cause := errors.New("original error")

				Convey("Seharusnya membungkus error standar menjadi AppError", func() {
					wrapped := WrapError(cause, ErrCodeInternalError, "wrapped")
					So(wrapped.Code, ShouldEqual, ErrCodeInternalError)
					So(wrapped.Message, ShouldEqual, "wrapped")
					So(errors.Is(wrapped, cause), ShouldBeTrue)
				})

				Convey("Seharusnya mengembalikan AppError yang sudah ada tanpa perubahan", func() {
					appErr := NewAppErrorWithCause(ErrCodeDatabaseError, "db error", cause)
					wrapped := WrapError(appErr, ErrCodeInternalError, "wrapped")
					So(wrapped, ShouldEqual, appErr)
					So(wrapped.Code, ShouldNotEqual, ErrCodeInternalError)
				})

				Convey("Seharusnya mengembalikan nil jika error yang diberikan nil", func() {
					wrapped := WrapError(nil, ErrCodeInternalError, "wrapped")
					So(wrapped, ShouldBeNil)
				})
			})
		})

		Convey("Fungsi-fungsi konstruktor", func() {
			Convey("NewValidationError", func() {
				fields := validator.ValidationErrors{
					validator.ValidationError{
						Field: "name",
						Tag:   "required",
					},
				}
				err := NewValidationError(fields)

				So(err, ShouldNotBeNil)
				So(err.Code, ShouldEqual, ErrCodeValidationFailed)
				So(err.Message, ShouldEqual, "Failed to validate request")
				So(err.Fields, ShouldResemble, fields)
			})

			Convey("NewBodyParserError", func() {
				err := NewBodyParserError()
				So(err, ShouldNotBeNil)
				So(err.Code, ShouldEqual, ErrCodeInvalidInput)
				So(err.Message, ShouldEqual, "Invalid request body")
			})
		})
	})
}
