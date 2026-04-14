package validation

// PaginateParams 分页参数
type PaginateParams struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
}

// Validate 分页参数校验
func (p PaginateParams) Validate() error {
	if p.Page < 1 || p.Page > 1000 {
		return ErrInvalidArgument("page 须在 [1, 1000] 范围内")
	}
	if p.Size < 1 || p.Size > 100 {
		return ErrInvalidArgument("size 须在 [1, 100] 范围内")
	}
	return nil
}
