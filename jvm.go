package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// 定义版本号
const version = "1.0.0"
const configFileName = "config.json"
const current = "current"

type Version struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Config struct {
	Path    string    `json:"path"`
	Version []Version `json:"version"`
}

func parseConfig() (Config, error) {
	path, _ := getExecutablePath(configFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	} else {
		return cfg, nil
	}
}

func help() {
	fmt.Printf("jvm version: %s\n", version)
	fmt.Println("Usage:")
	fmt.Println("  jvm list        # Lists all JDK versions.")
	fmt.Println("  jvm use jdk21   # Use a specific JDK version.")
	os.Exit(1)
}

func getExecutablePath(name string) (string, error) {
	// 获取可执行程序的全路径
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	// 获取可执行程序所在的目录，并拼接指定的文件名
	return filepath.Join(filepath.Dir(exePath), name), nil
}

func list(config Config) {
	// 即使出错也不处理异常，只是影响内容展示而已
	path, _ := getExecutablePath(current)
	data, _ := os.ReadFile(path)
	fmt.Println()
	for _, v := range config.Version {
		if data != nil && v.Name == string(data) {
			fmt.Printf("* %s\n", v.Name)
		} else {
			fmt.Printf("  %s\n", v.Name)
		}
	}
}

func use(config Config, name string) error {
	fmt.Printf("Switching...\n")
	// 先解析出源文件地址
	srcPath := ""
	for _, entry := range config.Version {
		if entry.Name == name {
			srcPath = entry.Path
		}
	}
	if srcPath == "" {
		return fmt.Errorf("%s not found", name)
	}

	// 将目录中的内容遍历删除
	dstPath := config.Path
	entries, err := os.ReadDir(dstPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entryPath := filepath.Join(dstPath, entry.Name())
		err = os.RemoveAll(entryPath)
		if err != nil {
			return err
		}
	}
	// 将内容复制到目录中
	err = copyDir(srcPath, dstPath)
	if err != nil {
		return err
	}
	fmt.Printf("Switched\n")
	fmt.Printf("Now using Java v%s\n", name)
	// 0644表示文件所有者可读写，其他人只读
	path, _ := getExecutablePath(current)
	os.WriteFile(path, []byte(name), 0644)
	return nil
}

func copyFile(srcFile, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	// 复制文件权限
	fi, err := src.Stat()
	if err != nil {
		return err
	}
	return os.Chmod(dstFile, fi.Mode())
}

func copyDir(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	// 确保目标目录存在
	err = os.MkdirAll(dstDir, 0755)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			// 递归复制子目录
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// 复制文件
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Printf("Failed to parse config: %v\n", err)
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		help()
	}
	if os.Args[1] == "list" {
		list(cfg)
		return
	} else if os.Args[1] == "use" && len(os.Args) == 3 {
		useName := os.Args[2]
		err = use(cfg, useName)
		if err != nil {
			fmt.Printf("Failed to use %s: %v\n", useName, err)
			os.Exit(1)
		}
		return
	} else {
		help()
	}

}
