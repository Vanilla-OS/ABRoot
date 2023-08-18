package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type PCSpecs struct {
	CPU    string
	GPU    []string
	Memory string
}

type GPUInfo struct {
	Address     string
	Description string
}

func getCPUInfo() (string, error) {
	info, err := cpu.Info()
	if err != nil {
		return "", err
	}
	if len(info) == 0 {
		return "", fmt.Errorf("CPU information not found")
	}
	return info[0].ModelName, nil
}

func parseGPUInfo(line string) (string, error) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("GPU information not found")
	}

	parts = strings.SplitN(parts[2], ":", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("GPU information not found")
	}

	return strings.TrimSpace(parts[1]), nil
}

func getGPUInfo() ([]string, error) {
	cmd := exec.Command("lspci")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error getting GPU info:", err)
		return nil, err
	}

	lines := strings.Split(string(output), "\n")

	var gpus []string
	for _, line := range lines {
		if strings.Contains(line, "VGA compatible controller") {
			gpu, err := parseGPUInfo(line)
			if err != nil {
				continue
			}
			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}

func getMemoryInfo() (string, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d MB", vm.Total/1024/1024), nil
}

func GetPCSpecs() PCSpecs {
	cpu, _ := getCPUInfo()
	gpu, _ := getGPUInfo()
	memory, _ := getMemoryInfo()

	return PCSpecs{
		CPU:    cpu,
		GPU:    gpu,
		Memory: memory,
	}
}
