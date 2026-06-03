package abstraction

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPaginationCursorLimit(t *testing.T) {
	tests := []struct {
		name     string
		cursor   *PaginationCursor
		expected int
	}{
		{
			name:     "nil cursor",
			cursor:   nil,
			expected: defaultPageSize,
		},
		{
			name:     "zero page size",
			cursor:   &PaginationCursor{PageSize: 0},
			expected: defaultPageSize,
		},
		{
			name:     "negative page size",
			cursor:   &PaginationCursor{PageSize: -10},
			expected: defaultPageSize,
		},
		{
			name:     "page size exceeds default",
			cursor:   &PaginationCursor{PageSize: 200},
			expected: defaultPageSize,
		},
		{
			name:     "valid page size below default",
			cursor:   &PaginationCursor{PageSize: 50},
			expected: 50,
		},
		{
			name:     "valid page size equals default",
			cursor:   &PaginationCursor{PageSize: 100},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cursor.Limit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationCursorGetPageSize(t *testing.T) {
	tests := []struct {
		name     string
		cursor   *PaginationCursor
		expected int
	}{
		{
			name:     "nil cursor",
			cursor:   nil,
			expected: defaultPageSize,
		},
		{
			name:     "zero page size",
			cursor:   &PaginationCursor{PageSize: 0},
			expected: defaultPageSize,
		},
		{
			name:     "negative page size",
			cursor:   &PaginationCursor{PageSize: -5},
			expected: defaultPageSize,
		},
		{
			name:     "page size exceeds default",
			cursor:   &PaginationCursor{PageSize: 150},
			expected: defaultPageSize,
		},
		{
			name:     "valid page size",
			cursor:   &PaginationCursor{PageSize: 25},
			expected: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cursor.GetPageSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationLimit(t *testing.T) {
	tests := []struct {
		name     string
		page     *Pagination
		expected int
	}{
		{
			name:     "nil pagination",
			page:     nil,
			expected: defaultPageSize,
		},
		{
			name:     "zero page size",
			page:     &Pagination{PageSize: 0},
			expected: defaultPageSize,
		},
		{
			name:     "negative page size",
			page:     &Pagination{PageSize: -10},
			expected: defaultPageSize,
		},
		{
			name:     "valid page size",
			page:     &Pagination{PageSize: 50},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.page.Limit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     *Pagination
		expected int
	}{
		{
			name:     "nil pagination",
			page:     nil,
			expected: 0,
		},
		{
			name:     "zero page",
			page:     &Pagination{Page: 0},
			expected: 0,
		},
		{
			name:     "negative page",
			page:     &Pagination{Page: -5},
			expected: 0,
		},
		{
			name:     "page 1",
			page:     &Pagination{Page: 1},
			expected: 0,
		},
		{
			name:     "page 2",
			page:     &Pagination{Page: 2},
			expected: 1,
		},
		{
			name:     "page 10",
			page:     &Pagination{Page: 10},
			expected: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.page.Offset()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationGetPage(t *testing.T) {
	tests := []struct {
		name     string
		page     *Pagination
		expected int
	}{
		{
			name:     "nil pagination",
			page:     nil,
			expected: 1,
		},
		{
			name:     "zero page",
			page:     &Pagination{Page: 0},
			expected: 1,
		},
		{
			name:     "negative page",
			page:     &Pagination{Page: -5},
			expected: 1,
		},
		{
			name:     "valid page",
			page:     &Pagination{Page: 5},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.page.GetPage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationGetPageSize(t *testing.T) {
	tests := []struct {
		name     string
		page     *Pagination
		expected int
	}{
		{
			name:     "nil pagination",
			page:     nil,
			expected: defaultPageSize,
		},
		{
			name:     "zero page size",
			page:     &Pagination{PageSize: 0},
			expected: defaultPageSize,
		},
		{
			name:     "negative page size",
			page:     &Pagination{PageSize: -5},
			expected: defaultPageSize,
		},
		{
			name:     "page size exceeds default",
			page:     &Pagination{PageSize: 200},
			expected: defaultPageSize,
		},
		{
			name:     "valid page size",
			page:     &Pagination{PageSize: 50},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.page.GetPageSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationGetSortBy(t *testing.T) {
	tests := []struct {
		name     string
		page     *Pagination
		expected string
	}{
		{
			name:     "nil pagination",
			page:     nil,
			expected: "",
		},
		{
			name:     "nil sort by and nil order by",
			page:     &Pagination{},
			expected: "id desc",
		},
		{
			name:     "nil sort by with asc order",
			page:     &Pagination{OrderBy: ptr("asc")},
			expected: "id asc",
		},
		{
			name:     "nil sort by with desc order",
			page:     &Pagination{OrderBy: ptr("desc")},
			expected: "id desc",
		},
		{
			name:     "sort by created_at",
			page:     &Pagination{SortBy: ptr("created_at"), OrderBy: ptr("asc")},
			expected: "created_at asc",
		},
		{
			name:     "sort by with order keyword",
			page:     &Pagination{SortBy: ptr("order"), OrderBy: ptr("desc")},
			expected: "\"order\" desc",
		},
		{
			name:     "sort by with invalid order defaults to desc",
			page:     &Pagination{SortBy: ptr("id"), OrderBy: ptr("invalid")},
			expected: "id desc",
		},
		{
			name:     "sort by with empty order defaults to desc",
			page:     &Pagination{SortBy: ptr("id"), OrderBy: ptr("")},
			expected: "id desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.page.GetSortBy()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationGetSorting(t *testing.T) {
	tests := []struct {
		name     string
		page     *Pagination
		expected *Sorting
	}{
		{
			name:     "nil sort by returns nil",
			page:     &Pagination{},
			expected: nil,
		},
		{
			name:     "with sort by and nil order by",
			page:     &Pagination{SortBy: ptr("id")},
			expected: &Sorting{SortBy: "id", OrderBy: "asc"},
		},
		{
			name:     "with sort by and order by",
			page:     &Pagination{SortBy: ptr("created_at"), OrderBy: ptr("desc")},
			expected: &Sorting{SortBy: "created_at", OrderBy: "desc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.page.GetSorting()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewSorting(t *testing.T) {
	tests := []struct {
		name     string
		sortBy   string
		sort     string
		expected *Sorting
	}{
		{
			name:     "asc order",
			sortBy:   "id",
			sort:     "asc",
			expected: &Sorting{SortBy: "id", OrderBy: "asc"},
		},
		{
			name:     "desc order",
			sortBy:   "created_at",
			sort:     "desc",
			expected: &Sorting{SortBy: "created_at", OrderBy: "desc"},
		},
		{
			name:     "invalid order defaults to asc",
			sortBy:   "id",
			sort:     "invalid",
			expected: &Sorting{SortBy: "id", OrderBy: "asc"},
		},
		{
			name:     "empty order defaults to asc",
			sortBy:   "id",
			sort:     "",
			expected: &Sorting{SortBy: "id", OrderBy: "asc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSorting(tt.sortBy, tt.sort)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPageInfo(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		moreRecords bool
		page        *Pagination
		sorting     *Sorting
		verify      func(t *testing.T, info *PaginationInfo)
	}{
		{
			name:        "with nil sorting",
			count:       10,
			moreRecords: true,
			page:        &Pagination{Page: 1, PageSize: 50},
			sorting:     nil,
			verify: func(t *testing.T, info *PaginationInfo) {
				assert.Equal(t, 10, info.Count)
				assert.Equal(t, true, info.MoreRecords)
				assert.Equal(t, 1, info.Page)
				assert.Equal(t, 50, info.PageSize)
				assert.NotNil(t, info.Sorting)
				assert.Equal(t, "", info.Sorting.SortBy)
				assert.Equal(t, "", info.Sorting.OrderBy)
			},
		},
		{
			name:        "with sorting",
			count:       20,
			moreRecords: false,
			page:        &Pagination{Page: 2, PageSize: 25},
			sorting:     &Sorting{SortBy: "id", OrderBy: "asc"},
			verify: func(t *testing.T, info *PaginationInfo) {
				assert.Equal(t, 20, info.Count)
				assert.Equal(t, false, info.MoreRecords)
				assert.Equal(t, 2, info.Page)
				assert.Equal(t, 25, info.PageSize)
				assert.Equal(t, "id", info.Sorting.SortBy)
				assert.Equal(t, "asc", info.Sorting.OrderBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPageInfo(tt.count, tt.moreRecords, tt.page, tt.sorting)
			tt.verify(t, result)
		})
	}
}

func TestPaginationApply(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	type TestModel struct {
		ID   int
		Name string
	}

	_ = db.AutoMigrate(&TestModel{})

	tests := []struct {
		name   string
		page   *Pagination
		verify func(t *testing.T, db *gorm.DB)
	}{
		{
			name: "nil pagination",
			page: nil,
			verify: func(t *testing.T, db *gorm.DB) {
				assert.NotNil(t, db)
			},
		},
		{
			name: "pagination without order by",
			page: &Pagination{
				Page:     1,
				PageSize: 50,
			},
			verify: func(t *testing.T, db *gorm.DB) {
				assert.NotNil(t, db)
			},
		},
		{
			name: "pagination with order by",
			page: &Pagination{
				Page:     1,
				PageSize: 50,
				SortBy:   ptr("id"),
				OrderBy:  ptr("asc"),
			},
			verify: func(t *testing.T, db *gorm.DB) {
				assert.NotNil(t, db)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := db.Session(&gorm.Session{})
			if tt.page != nil {
				tt.page.Apply(testDB)
			}
			tt.verify(t, testDB)
		})
	}
}

func ptr(s string) *string {
	return &s
}
