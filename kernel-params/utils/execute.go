package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetAllKernelParams() (string, error) {
	cmd := exec.Command("sysctl", "-a")
	out, err := cmd.Output()
	return string(out), err
}

func GetSingalKernelParams(name string) (string, error) {
	cmd := exec.Command("sysctl", name)
	out, err := cmd.Output()
	return string(out), err
}

func setKernelParams(config string) error {
	cmd := exec.Command("sysctl", "-w", config)
	_, err := cmd.Output()
	return err
}

func CompareKernelParams(expect map[string]string) {
	for k, v := range expect {
		c, err := GetSingalKernelParams(k)
		if err != nil {
			fmt.Println(err)
			continue
		}
		s := strings.TrimSpace(strings.Split(c, "=")[1])
		if s != v {
			fmt.Println(k, " except: ", v, "get:", s)
			config := k + "=" + v
			err := setKernelParams(config)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("%s get %s match expect\n", k, s)
		}
	}
}
