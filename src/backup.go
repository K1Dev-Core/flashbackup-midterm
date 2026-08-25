package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) handleMove(raw string) {
	if a.source == "" {
		fmt.Println("Warning: set a source with /source first.")
		return
	}
	if a.dest == "" {
		fmt.Println("Warning: set a destination with /dest first.")
		return
	}
	names, err := requestedNames(raw)
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	if len(names) == 1 && strings.EqualFold(names[0], "all") {
		entries, readErr := os.ReadDir(a.source)
		if readErr != nil {
			fmt.Println("Warning:", readErr)
			return
		}
		names = names[:0]
		for _, entry := range entries {
			if isBackupFile(entry) {
				names = append(names, entry.Name())
			}
		}
	}
	start := time.Now()
	startStr := start.Format("15:04:05.000")
	results := a.moveFilesWithProgress(names, startStr)
	finish := time.Now()
	finishStr := finish.Format("15:04:05.000")
	elapsed := float64(finish.Sub(start).Nanoseconds()) / 1e6
	moved := 0
	for _, result := range results {
		if result.err == nil {
			moved++
		}
	}
	fmt.Printf("ย้ายไฟล์แล้ว %d ไฟล์\n", moved)
	fmt.Printf("เริ่ม: %s\n", startStr)
	fmt.Printf("สิ้นสุด: %s\n", finishStr)
	fmt.Printf("ใช้เวลา: %.6fms\n", elapsed)
}

type moveResult struct {
	name string
	err  error
}

func (a *App) moveFilesWithProgress(names []string, startStr string) []moveResult {
	total := len(names)
	results := make([]moveResult, 0, total)
	existing := map[string]bool{}
	if total > 0 {
		var files []File
		if err := a.db.Where("dest = ? AND filename IN ?", a.dest, names).Find(&files).Error; err != nil {
			for _, name := range names {
				results = append(results, moveResult{name: name, err: fmt.Errorf("check database: %w", err)})
			}
			return results
		}
		for _, file := range files {
			existing[file.Filename] = true
		}
	}

	moved := make([]File, 0, total)
	movedNames := make([]string, 0, total)
	done := 0
	fmt.Printf("\r[%s] %d/%d %3d%% %-30s", strings.Repeat("-", 20), 0, total, 0, "")
	for _, name := range names {
		done++
		pct := done * 100 / total
		barW := pct * 20 / 100
		bar := strings.Repeat("#", barW) + strings.Repeat("-", 20-barW)
		if existing[name] {
			fmt.Printf("\r[%s] %d/%d %3d%% %-30s", bar, done, total, pct, name)
			results = append(results, moveResult{name: name, err: errors.New("file already exists in database history")})
			continue
		}
		if err := a.transferFile(name); err != nil {
			fmt.Printf("\r[%s] %d/%d %3d%% %-30s", bar, done, total, pct, name)
			results = append(results, moveResult{name: name, err: err})
			continue
		}
		moved = append(moved, File{Dest: a.dest, Filename: name})
		movedNames = append(movedNames, name)
		results = append(results, moveResult{name: name})
		fmt.Printf("\r[%s] %d/%d %3d%% %-30s", bar, done, total, pct, name)
	}
	fmt.Println()
	if len(moved) == 0 {
		return results
	}
	if err := a.db.Create(&moved).Error; err != nil {
		rollbackErrors := map[string]error{}
		for i := len(movedNames) - 1; i >= 0; i-- {
			name := movedNames[i]
			src := filepath.Join(a.source, name)
			dst := filepath.Join(a.dest, name)
			if rollbackErr := transfer(dst, src); rollbackErr != nil {
				rollbackErrors[name] = rollbackErr
			}
		}
		for i := range results {
			if results[i].err != nil {
				continue
			}
			if rollbackErr := rollbackErrors[results[i].name]; rollbackErr != nil {
				results[i].err = fmt.Errorf("save history: %v; rollback: %v", err, rollbackErr)
			} else {
				results[i].err = fmt.Errorf("save history: %w", err)
			}
		}
	}
	return results
}

func (a *App) transferFile(name string) error {
	if err := validFilename(name); err != nil {
		return err
	}
	src := filepath.Join(a.source, name)
	dst := filepath.Join(a.dest, name)
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("source is not a regular file")
	}
	if _, err := os.Stat(dst); err == nil {
		return errors.New("destination file already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check destination: %w", err)
	}
	if err := transfer(src, dst); err != nil {
		return err
	}
	return nil
}

func transfer(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy source: %w", err)
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		in.Close()
		return fmt.Errorf("copy destination: %w", err)
	}
	_, copyErr := io.Copy(out, in)
	closeOutErr := out.Close()
	closeInErr := in.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("copy: %w", copyErr)
	}
	if closeOutErr != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("close destination: %w", closeOutErr)
	}
	if closeInErr != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("close source: %w", closeInErr)
	}
	if err := os.Remove(src); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("remove source: %w", err)
	}
	return nil
}
