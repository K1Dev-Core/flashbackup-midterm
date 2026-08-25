package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) handleList(args []string) {
	if len(args) == 0 {
		fmt.Println("วิธีใช้: /list source | dest | db [path]")
		return
	}
	kind := strings.ToLower(args[0])
	path := ""
	if len(args) > 1 {
		path = strings.Join(args[1:], " ")
	}
	switch kind {
	case "source":
		if path == "" {
			path = a.source
		}
		a.listDirectory(path, "source")
	case "dest":
		if path == "" {
			path = a.dest
		}
		a.listDirectory(path, "dest")
	case "db":
		if path == "" {
			path = a.dest
		}
		path, err := directory(path)
		if err != nil {
			fmt.Println("Warning: destination is required for /list db.")
			return
		}
		var files []File
		if err := a.db.Where("dest = ?", path).Order("idx").Find(&files).Error; err != nil {
			fmt.Println("Warning:", err)
			return
		}
		if len(files) == 0 {
			fmt.Printf("ยังไม่มีประวัติของ dest: %s\n", path)
			return
		}
		fmt.Printf("ประวัติ dest: %s\n", path)
		fmt.Println("idx\tfilename")
		for _, file := range files {
			fmt.Printf("%d\t%s\n", file.Idx, file.Filename)
		}
	default:
		fmt.Println("วิธีใช้: /list source | dest | db [path]")
	}
}

func (a *App) listDirectory(raw, label string) {
	path, err := directory(raw)
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	names := make([]string, 0)
	for _, entry := range entries {
		if isBackupFile(entry) {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		fmt.Printf("ไม่มีไฟล์ใน %s: %s\n", label, path)
		return
	}
	fmt.Printf("%s: %s\n", label, path)
	for _, name := range names {
		fmt.Println(name)
	}
	fmt.Printf("(%d ไฟล์)\n", len(names))
}

func (a *App) check() {
	if a.dest == "" {
		fmt.Println("Warning: set a destination with /dest first.")
		return
	}
	a.integrity(a.dest)
}

func (a *App) integrity(dest string) {
	var files []File
	if err := a.db.Where("dest = ?", dest).Order("idx").Find(&files).Error; err != nil {
		fmt.Println("Warning:", err)
		return
	}
	if len(files) == 0 {
		return
	}
	missing := make([]string, 0)
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(dest, file.Filename)); errors.Is(err, os.ErrNotExist) {
			missing = append(missing, file.Filename)
		}
	}
	if len(missing) == 0 {
		fmt.Printf("ไฟล์ใน dest ตรงกับฐานข้อมูล (%d ไฟล์)\n", len(files))
		return
	}
	fmt.Println("ไฟล์ใน dest ไม่ตรงกับฐานข้อมูล")
	for _, name := range missing {
		fmt.Printf("มีในฐานข้อมูลแต่ไม่มีใน dest: %s\n", name)
	}
}
