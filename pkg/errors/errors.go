package errors

import "errors"

// 业务错误定义
var (
	ErrRecordNotFound = errors.New("记录不存在")
	ErrInvalidParam   = errors.New("参数无效")
	ErrSaveFailed     = errors.New("保存失败")
	ErrMatchFailed    = errors.New("匹配失败")
	ErrExportFailed   = errors.New("导出失败")
	ErrImportFailed   = errors.New("导入失败")
	ErrDuplicate      = errors.New("记录已存在")
)
