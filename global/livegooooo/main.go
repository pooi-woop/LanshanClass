// FilePath: C:/LanshanClass1.3/global/livegooooo\main.go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// 设置 LiveGo 二进制文件的路径
	livegoPath := "C:/woopapp/livego_0.0.20_windows_amd64/livego.exe"

	// 检查文件是否存在
	if _, err := os.Stat(livegoPath); os.IsNotExist(err) {
		fmt.Printf("LiveGo executable not found at %s\n", livegoPath)
		return
	}

	// 启动 LiveGo 服务器
	startCmd := exec.Command(livegoPath)
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	// 启动服务器
	err := startCmd.Start()
	if err != nil {
		fmt.Printf("Failed to start LiveGo server: %v\n", err)
		return
	}

	fmt.Println("LiveGo server started successfully. You can now push and pull streams.")

}
