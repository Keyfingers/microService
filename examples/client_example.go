package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// 这是一个示例客户端程序，展示如何调用微服务的各个接口

const baseURL = "http://localhost:8080"

func main() {
	fmt.Println("=== 微服务客户端示例 ===")
	fmt.Println()

	// 1. 健康检查
	fmt.Println("1. 健康检查...")
	healthCheck()
	fmt.Println()

	// 2. 详细健康检查
	fmt.Println("2. 详细健康检查...")
	detailedHealthCheck()
	fmt.Println()

	// 3. 发送消息到队列
	fmt.Println("3. 发送消息到队列...")
	sendMessage()
	fmt.Println()

	// 4. 上传文件（需要提供文件路径）
	// 取消注释以下代码并提供实际文件路径
	// fmt.Println("4. 上传文件...")
	// uploadFile("/path/to/your/file.jpg")
	// fmt.Println()

	fmt.Println("=== 示例完成 ===")
}

// healthCheck 基础健康检查
func healthCheck() {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}

// detailedHealthCheck 详细健康检查
func detailedHealthCheck() {
	resp, err := http.Get(baseURL + "/health/detail")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}

// sendMessage 发送消息到队列
func sendMessage() {
	message := map[string]interface{}{
		"queue": "task",
		"message": map[string]interface{}{
			"type":    "send_email",
			"to":      "user@example.com",
			"subject": "测试邮件",
			"body":    "这是一条测试消息",
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}

	resp, err := http.Post(
		baseURL+"/api/v1/message",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}

// uploadFile 上传文件
func uploadFile(filePath string) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer file.Close()

	// 创建 multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Printf("创建表单失败: %v\n", err)
		return
	}

	// 复制文件内容
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("复制文件失败: %v\n", err)
		return
	}

	// 关闭 writer
	writer.Close()

	// 发送请求
	req, err := http.NewRequest("POST", baseURL+"/api/v1/upload", body)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(respBody))
}
