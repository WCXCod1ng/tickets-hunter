package xerr

// IsServerError 判断一个错误码是否属于服务器端错误，兼容go-zero的错误码
func IsServerError(code uint32) bool {

	if code >= 7 && code <= 15 {
		return true
	} else if code == 1 || code == 2 || code == 4 {
		return true
	}

	return false
}
