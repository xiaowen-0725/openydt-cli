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

// StatusHint returns an actionable, AI-agent-friendly suggestion for a status.
func StatusHint(s int) string {
	switch s {
	case StatusSuccess:
		return ""
	case StatusBizFail:
		return "业务校验未通过, 见 resultCode 与 message; 多为入参取值/数据状态问题, 用 `openydt schema <cmd>` 查参数与可选值"
	case StatusSysError:
		return "服务端异常, 可稍后重试; 若持续请联系平台"
	case StatusSignError:
		return "签名错误: 核对 key/secret、时间(yyyyMMddHHmmss, 10分钟内有效)、--sign 版本(默认 v2); 测试环境通常仅接受 v2"
	case StatusKeyError:
		return "key 错误: 检查 --profile 对应的 key/secret 是否正确(openydt config list)"
	case StatusNoAuth:
		return "未授权: 该 key 无权访问此接口/车场; 用 `openydt park get-auth-park-codes` 查授权车场"
	case StatusBadParams:
		return "请求参数不完整: 用 `openydt schema <cmd>` 查必填参数, 或加 --body 补齐"
	case 9:
		return "接口在当前环境未部署; 换环境(--env)或确认平台已开通该接口"
	default:
		return ""
	}
}

// ResultHint returns an actionable suggestion for a business result code.
func ResultHint(code int) string {
	switch code {
	case 903:
		return "确认运营商/授权商身份与 key 是否匹配"
	case 904, 910:
		return "parkCode 不在授权车场内; 用 `openydt park get-auth-park-codes` 查可用车场"
	case 905, 1801:
		return "该车不在场或不存在; 用 `openydt parking get-park-on-site-car` 查在场车辆"
	case 906:
		return "账单不存在; 先 `openydt trade get-park-fee` 查费获取 parkingCode/账单"
	case 907:
		return "账单已同步, 无需重复操作"
	case 909:
		return "参数错误; 用 `openydt schema <cmd>` 核对参数名/必填/类型/可选值"
	case 911:
		return "无权操作该车场; 确认该 parkCode 在授权范围内"
	case 912:
		return "查费已超时(>10分钟), 请重新 get-park-fee 后再缴费"
	default:
		return ""
	}
}

// Retriable reports whether a status is worth retrying (transient).
func Retriable(status int) bool { return status == StatusSysError }

