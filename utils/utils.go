package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// 判断字符串 tg 是否能在字符串类型的切片 source 中找到
func Contains(source []string, tg string) bool {
	// 使用 range 关键字获取每个元素的索引和值，
	// 并分别赋值给变量 _（省略符号）和变量 s
	for _, s := range source {
		if s == tg {
			return true
		}
	}
	return false
}

// 用哈希算法生成一个唯一的 uuid
func Md5String(s string) string {
	// 创建一个 md5 实例
	h := md5.New()

	// 通过调用 h.Write([]byte(s)) 将字符串 s 的字节表示写入到 md5 实例中
	h.Write([]byte(s))

	// 调用 h.Sum(nil) 完成 MD5 的计算并返回结果
	// Sum() 方法接受一个切片参数，用于存储 MD5 值
	// 由于我们在这里不需要对 MD5 值进行其他操作，所以将参数设置为 nil
	// 再将计算得到的 MD5 值以十六进制的形式表示，并由 hex.EncodeToString() 转换为字符串
	// 注意：
	// 这段代码的作用是对给定的字符串进行 MD5 计算，并将计算结果以字符串的形式返回
	// 通常，MD5 值用于验证数据的完整性和唯一性，但请注意，MD5 算法已经被认为不再安全，不适用于密码等敏感数据的加密
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

// 生成一个新的 session 字符串
func GenerateSession(userName string) string {
	// 使用 fmt.Sprintf() 函数将用户名和字符串 "session" 拼接为一个新的字符串
	// 将拼接得到的新字符串作为参数传递给 Md5String() 函数，通过哈希算法生成 MD5 值的字符串表示
	return Md5String(fmt.Sprintf("%s:%s", userName, "session"))
}
