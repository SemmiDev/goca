package request

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilter(t *testing.T) {
	Convey("Testing Filter struct and methods", t, func() {
		Convey("NewFilter should return a filter with default values", func() {
			f := NewFilter()
			So(f.CurrentPage, ShouldEqual, DefaultCurrentPage)
			So(f.PerPage, ShouldEqual, DefaultPageLimit)
			So(f.SortDirection, ShouldEqual, DefaultColumnDirection)
		})

		Convey("GetLimit should return PerPage", func() {
			f := Filter{PerPage: 50}
			So(f.GetLimit(), ShouldEqual, 50)
		})

		Convey("GetOffset should calculate correctly", func() {
			f := Filter{PerPage: 10}

			f.CurrentPage = 1
			So(f.GetOffset(), ShouldEqual, 0)

			f.CurrentPage = 2
			So(f.GetOffset(), ShouldEqual, 10)

			f.CurrentPage = 5
			So(f.GetOffset(), ShouldEqual, 40)
		})

		Convey("HasKeyword should work correctly", func() {
			f := Filter{Keyword: "search term"}
			So(f.HasKeyword(), ShouldBeTrue)

			f.Keyword = ""
			So(f.HasKeyword(), ShouldBeFalse)
		})

		Convey("HasSort should work correctly", func() {
			f := Filter{SortBy: "created_at"}
			So(f.HasSort(), ShouldBeTrue)

			f.SortBy = ""
			So(f.HasSort(), ShouldBeFalse)
		})

		Convey("IsDesc should be case-insensitive", func() {
			f := Filter{}

			f.SortDirection = "desc"
			So(f.IsDesc(), ShouldBeTrue)

			f.SortDirection = "DESC"
			So(f.IsDesc(), ShouldBeTrue)

			f.SortDirection = "Desc"
			So(f.IsDesc(), ShouldBeTrue)

			f.SortDirection = "asc"
			So(f.IsDesc(), ShouldBeFalse)

			f.SortDirection = ""
			So(f.IsDesc(), ShouldBeFalse)
		})

		Convey("IsUnlimitedPage should work correctly", func() {
			f := Filter{PerPage: UnlimitedPage}
			So(f.IsUnlimitedPage(), ShouldBeTrue)

			f.PerPage = 20
			So(f.IsUnlimitedPage(), ShouldBeFalse)
		})
	})
}

func TestNewPaging(t *testing.T) {
	Convey("Testing NewPaging function", t, func() {

		Convey("Standard cases - Happy path", func() {
			Convey("When on the first page", func() {
				p, err := NewPaging(1, 10, 100)
				So(err, ShouldBeNil)
				So(p.HasPreviousPage, ShouldBeFalse)
				So(p.HasNextPage, ShouldBeTrue)
				So(p.CurrentPage, ShouldEqual, 1)
				So(p.PerPage, ShouldEqual, 10)
				So(p.TotalData, ShouldEqual, 100)
				So(p.TotalDataInCurrentPage, ShouldEqual, 10)
				So(p.LastPage, ShouldEqual, 10)
				So(p.From, ShouldEqual, 1)
				So(p.To, ShouldEqual, 10)
			})

			Convey("When on a middle page", func() {
				p, err := NewPaging(5, 10, 100)
				So(err, ShouldBeNil)
				So(p.HasPreviousPage, ShouldBeTrue)
				So(p.HasNextPage, ShouldBeTrue)
				So(p.CurrentPage, ShouldEqual, 5)
				So(p.LastPage, ShouldEqual, 10)
				So(p.From, ShouldEqual, 41)
				So(p.To, ShouldEqual, 50)
			})

			Convey("When on the last page (partial data)", func() {
				p, err := NewPaging(10, 10, 93)
				So(err, ShouldBeNil)
				So(p.HasPreviousPage, ShouldBeTrue)
				So(p.HasNextPage, ShouldBeFalse)
				So(p.CurrentPage, ShouldEqual, 10)
				So(p.LastPage, ShouldEqual, 10)
				So(p.TotalDataInCurrentPage, ShouldEqual, 3)
				So(p.From, ShouldEqual, 91)
				So(p.To, ShouldEqual, 93)
			})
		})

		Convey("Edge cases", func() {
			Convey("When totalData is 0", func() {
				p, err := NewPaging(1, 10, 0)
				So(err, ShouldBeNil)
				So(p.TotalData, ShouldEqual, 0)
				So(p.TotalDataInCurrentPage, ShouldEqual, 0)
				So(p.CurrentPage, ShouldEqual, 1)
				So(p.LastPage, ShouldEqual, 1)
				So(p.From, ShouldEqual, 0)
				So(p.To, ShouldEqual, 0)
				So(p.HasNextPage, ShouldBeFalse)
				So(p.HasPreviousPage, ShouldBeFalse)
			})

			Convey("When perPage is unlimited", func() {
				p, err := NewPaging(1, UnlimitedPage, 123)
				So(err, ShouldBeNil)
				So(p.PerPage, ShouldEqual, UnlimitedPage)
				So(p.TotalData, ShouldEqual, 123)
				So(p.TotalDataInCurrentPage, ShouldEqual, 123)
				So(p.LastPage, ShouldEqual, 1)
				So(p.From, ShouldEqual, 1)
				So(p.To, ShouldEqual, 123)
			})

			Convey("When totalData is less than perPage", func() {
				p, err := NewPaging(1, 20, 15)
				So(err, ShouldBeNil)
				So(p.LastPage, ShouldEqual, 1)
				So(p.HasNextPage, ShouldBeFalse)
				So(p.From, ShouldEqual, 1)
				So(p.To, ShouldEqual, 15)
				So(p.TotalDataInCurrentPage, ShouldEqual, 15)
			})

			Convey("When currentPage is out of bounds (too high)", func() {
				p, err := NewPaging(99, 10, 50) // Minta halaman 99, padahal hanya ada 5 halaman
				So(err, ShouldBeNil)
				So(p.CurrentPage, ShouldEqual, 5) // Seharusnya dikoreksi ke halaman terakhir
				So(p.LastPage, ShouldEqual, 5)
				So(p.HasNextPage, ShouldBeFalse)
				So(p.From, ShouldEqual, 0)
				So(p.To, ShouldEqual, 50)
			})
		})

		Convey("Error cases - Invalid input", func() {
			Convey("When perPage is 0", func() {
				p, err := NewPaging(1, 0, 100)
				So(p, ShouldBeNil)
				So(err, ShouldEqual, ErrPaging)
			})

			Convey("When perPage is negative (but not UnlimitedPage)", func() {
				p, err := NewPaging(1, -5, 100)
				So(p, ShouldBeNil)
				So(err, ShouldEqual, ErrPaging)
			})

			Convey("When currentPage is less than 1", func() {
				p, err := NewPaging(0, 10, 100) // Menyebabkan offset negatif
				So(p, ShouldBeNil)
				So(err, ShouldEqual, ErrPaging)
			})
		})
	})
}
