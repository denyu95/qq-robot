package model

import "time"

type Bill struct {
	Id          int       `json:"id"`
	Event       string    `json:"event"`
	Consumption float32   `json:"consumption"`
	SysDate     time.Time `json:"sysDate"`
	UpdateDate  time.Time `json:"updateDate"`
	Uid         string    `json:"uid"`
}
