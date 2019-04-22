// @APIVersion 1.0.0
// @Title 数据库查询分页模型
// @Description 数据库查询分页模型，用户需要分页进行查询的接口
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package models

const (
	MaxQueryNumber = 50
	MinQueryNumber = 1
	MinQueryPage   = 1
)

// PagingInfo 分页接口
type PagingInfo interface {
	Offset() int
	Limit() int
}

// Paging 分页类型
type Paging struct {
	offset int
	limit  int
}

// NewPaging 新的分页对象
func NewPaging(page, limit int) *Paging {
	return &Paging{
		offset: (page - 1) * limit,
		limit:  limit,
	}
}

// Offset 偏移量
func (p *Paging) Offset() int {
	return p.offset
}

// Limit 每页的数量
func (p *Paging) Limit() int {
	return p.limit
}
