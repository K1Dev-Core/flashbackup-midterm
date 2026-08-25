package main

type File struct {
	Idx      uint   `gorm:"column:idx;primaryKey;autoIncrement"`
	Dest     string `gorm:"column:dest;not null"`
	Filename string `gorm:"column:filename;not null"`
}

func (File) TableName() string { return "files" }

type Setting struct {
	Idx    uint   `gorm:"column:idx;primaryKey;autoIncrement"`
	Source string `gorm:"column:source;not null"`
	Dest   string `gorm:"column:dest;not null"`
}

func (Setting) TableName() string { return "settings" }
