package main

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (a *App) loadSettings() {
	var setting Setting
	if err := a.db.Order("idx DESC").First(&setting).Error; err == nil {
		if isDir(setting.Source) {
			a.source = setting.Source
		}
		if isDir(setting.Dest) {
			a.dest = setting.Dest
		}
	}
}

func (a *App) saveSettings() error {
	if a.source == "" || a.dest == "" {
		return nil
	}
	var setting Setting
	if err := a.db.Order("idx DESC").First(&setting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return a.db.Create(&Setting{Source: a.source, Dest: a.dest}).Error
	} else if err == nil {
		return a.db.Model(&setting).Updates(map[string]any{"source": a.source, "dest": a.dest}).Error
	} else {
		return err
	}
}

func (a *App) lastSettingIdx() uint {
	var setting Setting
	if err := a.db.Order("idx DESC").First(&setting).Error; err == nil {
		return setting.Idx
	}
	return 0
}

func (a *App) handleSource(raw string) {
	path, err := directory(raw)
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	a.source = path
	if err := a.saveSettings(); err != nil {
		fmt.Println("Warning:", err)
		return
	}
	idx := a.lastSettingIdx()
	fmt.Printf("บันทึก source แล้ว idx=%d source=%s\n", idx, path)
}

func (a *App) handleDest(raw string) {
	path, err := directory(raw)
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	a.dest = path
	if err := a.saveSettings(); err != nil {
		fmt.Println("Warning:", err)
		return
	}
	idx := a.lastSettingIdx()
	fmt.Printf("บันทึก dest แล้ว idx=%d dest=%s\n", idx, path)
	a.integrity(path)
}

func (a *App) handleSet(args []string) {
	if len(args) < 2 {
		fmt.Println("วิธีใช้: /set <source> <dest>")
		return
	}
	src, err := directory(args[0])
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	dst, err := directory(args[1])
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	a.source = src
	a.dest = dst
	if err := a.saveSettings(); err != nil {
		fmt.Println("Warning:", err)
		return
	}
	idx := a.lastSettingIdx()
	fmt.Printf("บันทึกตั้งค่าแล้ว idx=%d source=%s dest=%s\n", idx, src, dst)
}

func (a *App) handleAdd(args []string) {
	if len(args) < 2 {
		fmt.Println("วิธีใช้: /add <dest> <filename>")
		return
	}
	dst, err := directory(args[0])
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	filename := args[1]
	file := File{Dest: dst, Filename: filename}
	if err := a.db.Create(&file).Error; err != nil {
		fmt.Println("Warning:", err)
		return
	}
	fmt.Printf("บันทึกแล้ว idx=%d dest=%s filename=%s\n", file.Idx, dst, filename)
}

func (a *App) handleSettings() {
	var settings []Setting
	if err := a.db.Order("idx").Find(&settings).Error; err != nil {
		fmt.Println("Warning:", err)
		return
	}
	if len(settings) == 0 {
		fmt.Println("ยังไม่มีข้อมูลตั้งค่า")
		return
	}
	fmt.Println("idx\tsource\tdest")
	for _, s := range settings {
		fmt.Printf("%d\t%s\t%s\n", s.Idx, s.Source, s.Dest)
	}
}

func (a *App) handleClean() {
	a.db.Exec("DELETE FROM files")
	a.db.Exec("DELETE FROM settings")
	a.source = ""
	a.dest = ""
	fmt.Println("ลบข้อมูลในตาราง files และ settings แล้ว")
}
