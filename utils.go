package chatgpt

import "runtime"

func controlA() string {
	// 判断系统
	if runtime.GOOS == "darwin" {
		return "Meta+A"
	}
	return "Control+A"
}

func controlV() string {
	// 判断系统
	if runtime.GOOS == "darwin" {
		return "Meta+V"
	}
	return "Control+V"
}
