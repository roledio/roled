package models

type ResponseBody struct {
	Success    bool        `json:"success"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Error      *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Debug   *string `json:"debug,omitempty"`
}

type Pagination struct {
	PageNum   int `json:"page_num"`
	PageSize  int `json:"page_size"`
	TotalData int `json:"total_data"`
}
