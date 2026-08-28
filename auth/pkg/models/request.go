package models

type PageRequest struct {
	PageNum  int    `query:"page_num"`
	PageSize int    `query:"page_size"`
	SortBy   string `query:"sort_by"`
	SortDir  string `query:"sort_dir"`
}

// SetDefaults sets default values for page number and page size
func (pr *PageRequest) SetDefaults() {
	if pr.PageNum < 1 {
		pr.PageNum = 1
	}
	if pr.PageSize < 1 {
		pr.PageSize = 10
	}
	if pr.PageSize > 100 {
		pr.PageSize = 100
	}
	if pr.SortDir == "" {
		pr.SortDir = "asc"
	}
}

func (pr *PageRequest) Offset() int {
	return (pr.PageNum - 1) * pr.PageSize
}

func (pr *PageRequest) Limit() int {
	return pr.PageSize
}
