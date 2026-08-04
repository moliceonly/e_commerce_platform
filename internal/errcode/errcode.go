package errcode

// 统一业务错误码（阶段 H · 3.1）。
// HTTP 状态与业务 code 分离：Fail(c, httpStatus, errcode.Xxx, msg)

const (
	OK = 0

	// 通用
	ErrBadRequest   = 40000
	ErrUnauthorized = 40100
	ErrForbidden    = 40300
	ErrNotFound     = 40400
	ErrConflict     = 40900
	ErrInternal     = 50000

	// TODO(H1): 按域细分，例如
	// ErrAuthInvalidCreds = 40101
	// ErrAuthRateLimited  = 40102
	// ErrStockNotEnough   = 40901
)
