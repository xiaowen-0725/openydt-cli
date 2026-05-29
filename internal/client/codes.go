package client

// Business status codes (response "status" field).
const (
	StatusSuccess    = 1 // 业务成功
	StatusBizFail    = 2 // 业务失败 (see ResultCode)
	StatusSysError   = 3 // 系统异常
	StatusSignError  = 4 // 签名错误
	StatusKeyError   = 5 // key 错误
	StatusNoAuth     = 6 // 未授权
	StatusBadParams  = 7 // 请求参数不完整
)

// StatusText describes a status code.
func StatusText(s int) string {
	switch s {
	case StatusSuccess:
		return "业务成功"
	case StatusBizFail:
		return "业务失败"
	case StatusSysError:
		return "系统异常"
	case StatusSignError:
		return "签名错误"
	case StatusKeyError:
		return "key错误"
	case StatusNoAuth:
		return "未授权"
	case StatusBadParams:
		return "请求参数不完整"
	case 9:
		return "接口不存在"
	default:
		return "未知状态"
	}
}

// ResultText describes a business result code (valid when status == 2).
func ResultText(code int) string {
	switch code {
	case 0:
		return ""
	case 901:
		return "系统发生异常"
	case 902:
		return "远程服务器未响应"
	case 903:
		return "运营商不存在"
	case 904:
		return "停车场不存在"
	case 905:
		return "未找到在场车辆"
	case 906:
		return "账单不存在"
	case 907:
		return "账单已同步"
	case 908:
		return "其它错误"
	case 909:
		return "请求参数错误"
	case 910:
		return "找不到授权商下面的停车场"
	case 911:
		return "无权限操作该停车场"
	case 912:
		return "查费已超时，请重新查费"
	case 1801:
		return "找不到指定车辆"
	default:
		return "业务码未登记"
	}
}
