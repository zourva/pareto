package box

import (
	"os/exec"
	"regexp"
)

// CpuId returns the cpu identity of current host.
// The identity is retrieved by executing 'wmic cpu get ProcessorID'.
func CpuId() string {
	out, err := exec.Command("wmic", "cpu", "get", "ProcessorID").CombinedOutput()
	if err != nil {
		return ""
	}

	str := string(out)

	//匹配一个或多个空白符的正则表达式
	reg := regexp.MustCompile("\\s+")
	str = reg.ReplaceAllString(str, "")
	return str[11:]
}
