package request

import (
	"errors"
	"strings"
)

const (
	AscDirection           string = "asc"
	DescDirection          string = "desc"
	DefaultPageLimit       int    = 20
	DefaultCurrentPage     int    = 1
	DefaultColumnDirection string = AscDirection
	UnlimitedPage          int    = -1
)

type Filter struct {
	CurrentPage   int    `json:"current_page" form:"current_page" query:"current_page"`       // current page (berpindah-pindah halaman)
	PerPage       int    `json:"per_page" form:"per_page" query:"per_page"`                   // limit (batas data yang ditampilkan)
	Keyword       string `json:"keyword" form:"keyword" query:"keyword"`                      // search keyword (keyword pencarian)
	SortBy        string `json:"sort_by" form:"sort_by" query:"sort_by"`                      // column name to sort
	SortDirection string `json:"sort_direction" form:"sort_direction" query:"sort_direction"` // asc or desc direction
}

func NewFilter() Filter {
	return Filter{
		CurrentPage:   DefaultCurrentPage,
		PerPage:       DefaultPageLimit,
		SortDirection: DefaultColumnDirection,
	}
}

func (f *Filter) GetLimit() int {
	return f.PerPage
}

func (f *Filter) GetOffset() int {
	// ex:
	// page 1 -> (1 - 1) * 10 = 0
	// page 2 -> (2 - 1) * 10 = 10
	offset := (f.CurrentPage - 1) * f.PerPage
	return offset
}

func (f *Filter) HasKeyword() bool {
	return f.Keyword != ""
}

func (f *Filter) HasSort() bool {
	return f.SortBy != ""
}

func (f *Filter) IsDesc() bool {
	return strings.EqualFold(f.SortDirection, DescDirection)
}

func (f *Filter) IsUnlimitedPage() bool {
	return isUnlimitedPage(f.PerPage)
}

func isUnlimitedPage(perPage int) bool {
	return perPage == UnlimitedPage
}

type Paging struct {
	HasPreviousPage        bool `json:"has_previous_page"`
	HasNextPage            bool `json:"has_next_page"`
	CurrentPage            int  `json:"current_page"`
	PerPage                int  `json:"per_page"`
	TotalData              int  `json:"total_data"`
	TotalDataInCurrentPage int  `json:"total_data_in_current_page"`
	LastPage               int  `json:"last_page"`
	From                   int  `json:"from"`
	To                     int  `json:"to"`
}

var ErrPaging = errors.New("per_page harus lebih besar dari 0 dan offset tidak boleh negatif")

func NewPaging(currentPage, perPage, totalData int) (*Paging, error) {
	if isUnlimitedPage(perPage) {
		return &Paging{
			CurrentPage:            currentPage,
			PerPage:                perPage,
			TotalData:              totalData,
			LastPage:               1,
			From:                   1,
			To:                     totalData,
			TotalDataInCurrentPage: totalData,
		}, nil
	}

	if totalData == 0 {
		return &Paging{
			HasPreviousPage:        false,
			HasNextPage:            false,
			CurrentPage:            1,
			PerPage:                perPage,
			TotalData:              0,
			TotalDataInCurrentPage: 0,
			LastPage:               1,
			From:                   0,
			To:                     0,
		}, nil
	}

	offset := (currentPage - 1) * perPage

	if perPage <= 0 || offset < 0 {
		return nil, ErrPaging
	}

	lastPage := totalData / perPage
	if totalData%perPage != 0 {
		lastPage++
	}

	to := min(offset+perPage, totalData)
	from := int(0)
	if to > offset {
		from = offset + 1
	}

	if currentPage > lastPage {
		currentPage = lastPage
	}

	totalDataInCurrentPage := to - offset

	return &Paging{
		HasPreviousPage:        currentPage > 1,
		HasNextPage:            currentPage < lastPage,
		CurrentPage:            currentPage,
		PerPage:                perPage,
		TotalData:              totalData,
		LastPage:               lastPage,
		From:                   from,
		To:                     to,
		TotalDataInCurrentPage: totalDataInCurrentPage,
	}, nil
}
