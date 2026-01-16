// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package model

const TableNameCasbinRuleM = "casbin_rule"

// CasbinRuleM mapped from table <casbin_rule>
type CasbinRuleM struct {
	ID    uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PType string `gorm:"column:ptype" json:"ptype"`
	V0    string `gorm:"column:v0" json:"v0"`
	V1    string `gorm:"column:v1" json:"v1"`
	V2    string `gorm:"column:v2" json:"v2"`
	V3    string `gorm:"column:v3" json:"v3"`
	V4    string `gorm:"column:v4" json:"v4"`
	V5    string `gorm:"column:v5" json:"v5"`
}

// TableName CasbinRuleM's table name
func (*CasbinRuleM) TableName() string {
	return TableNameCasbinRuleM
}
