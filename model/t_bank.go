package model

import "time"

type Bank struct {
	Id		int       `json:"id"`
	Uid		string    `json:"uid"`
	Money	float32	  `json:"name"`
	SysDate	time.Time `json:"sysDate"`
}