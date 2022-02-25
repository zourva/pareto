package box

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

// CpuId returns the cpu identity of current host.
// The identity is retrieved by executing 'dmidecode -t processor|grep ID|head -1'.
func CpuId() string {
	cmd := exec.Command("/bin/sh", "-c", `dmidecode -t processor|grep ID|head -1`)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("cmd.StdoutPipe failed:", err)
		return ""
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("cmd.StderrPipe failed:", err)
		return ""
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("cmd.Start failed: ", err)
		return ""
	}

	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		fmt.Println("ReadAll stderr: ", err)
		return ""
	}

	if len(bytesErr) != 0 {
		fmt.Printf("stderr occurred: %s", string(bytesErr))
		return ""
	}

	bytesOut, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Println("ReadAll stdout: ", err)
		return ""
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("cmd.Wait failed: ", err)
		return ""
	}

	cpuId := string(bytesOut)
	cpuId = strings.Replace(cpuId, "ID: ", "", -1)
	cpuId = strings.Replace(cpuId, "\t", "", -1)
	cpuId = strings.Replace(cpuId, "\n", "", -1)
	cpuId = strings.Replace(cpuId, " ", "", -1)

	return cpuId
}
