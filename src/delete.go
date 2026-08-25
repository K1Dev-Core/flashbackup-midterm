package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (a *App) handleDelete(args []string) {
	if a.dest == "" {
		fmt.Println("Warning: set a destination with /dest first.")
		return
	}
	if len(args) < 2 || strings.ToLower(args[0]) != "dest" {
		fmt.Println("วิธีใช้: /delete <file>,... หรือ all")
		return
	}
	names, err := requestedNames(strings.Join(args[1:], " "))
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	if len(names) == 1 && strings.EqualFold(names[0], "all") {
		names = a.destinationNames()
	}
	deleted := 0
	for _, name := range names {
		if err := a.deleteOne(name); err != nil {
			fmt.Printf("ไม่พบไฟล์ใน dest: dest %s\n", name)
			continue
		}
		deleted++
	}
	fmt.Printf("ลบไฟล์แล้ว %d ไฟล์ และลบจากฐานข้อมูลแล้ว\n", deleted)
}

func (a *App) destinationNames() []string {
	seen := map[string]bool{}
	var files []File
	_ = a.db.Where("dest = ?", a.dest).Find(&files).Error
	for _, file := range files {
		seen[file.Filename] = true
	}
	if entries, err := os.ReadDir(a.dest); err == nil {
		for _, entry := range entries {
			if isBackupFile(entry) {
				seen[entry.Name()] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (a *App) deleteOne(name string) error {
	if err := validFilename(name); err != nil {
		return err
	}
	path := filepath.Join(a.dest, name)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove file: %w", err)
	}
	a.db.Where("dest = ? AND filename = ?", a.dest, name).Delete(&File{})
	return nil
}
