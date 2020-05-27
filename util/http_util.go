/**@description	HTTP相关的方法集合
@author pujie
**/
package util

import (
	"math/rand"
	"net"
	"strings"
)

// GetLocalIpByTcp 有外网的情况下, 通过tcp访问获得本机ip地址
func GetLocalIpByTcp() string {
	conn, err := net.Dial("tcp", "www.baidu.com:80")
	if err != nil {
		return ""
	}
	defer func() {
		_ = conn.Close()
	}()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}

// RandomInt 随机数
func RandomInt(num int) int {
	return rand.Intn(65536) % num
}
