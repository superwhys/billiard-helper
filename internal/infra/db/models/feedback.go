// File:		feedback.go
// Created by:	Hoven
// Created on:	2026-02-16
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package models

import "gorm.io/gorm"

type Feedback struct {
	gorm.Model
	UserID  uint   `gorm:"column:user_id;index;not null;comment:用户ID" json:"user_id"`
	Content string `gorm:"column:content;type:text;comment:反馈内容"`
}
