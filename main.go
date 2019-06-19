package main

import (
	"time"
	"regexp"

	"github.com/denyu95/qq-robot/model"
	"github.com/denyu95/qq-robot/dbutil"
	"github.com/juzi5201314/cqhttp-go-sdk/server"
	"github.com/golang/glog"
	"fmt"
	"strings"
)

func main() {
	s := server.StartListenServer(8080, "/")
	s.ListenGroupMessage(server.GroupMessageListener(group))
	s.Listen()
}

func group(m map[string]interface{}) map[string]interface{} {
	helpExp := regexp.MustCompile(`^帮助$`)
	getBillsExp := regexp.MustCompile(`^查看$`)
	deleteBillExp := regexp.MustCompile(`^删除((?:\s(?:\d+))+)$`)
	addBillExp := regexp.MustCompile(`^(?:！|!)([^\n]+)\s(\d+\.{0,1}\d{0,2})`)
	updateBillExp := regexp.MustCompile(`^编辑\s(\d+)\s([^\n]+)\s(\d+\.{0,1}\d{0,2})`)

	msg := m["message"].(string)

	if addBillExp.Match([]byte(msg)) {
		result := addBillExp.FindAllStringSubmatch(msg, -1)
		event := result[0][1]
		consumption := result[0][2]
		return addBill(event, consumption)

	} else if getBillsExp.Match([]byte(msg)) {
		return getBills()

	} else if deleteBillExp.Match([]byte(msg)) {
		result := deleteBillExp.FindAllStringSubmatch(msg, -1)
		ids := result[0][1]
		return deleteBill(ids)

	} else if updateBillExp.Match([]byte(msg)) {
		result := updateBillExp.FindAllStringSubmatch(msg, -1)
		id := result[0][1]
		event := result[0][2]
		consumption := result[0][3]
		return updateBill(id, event, consumption)

	} else if helpExp.Match([]byte(msg)) {
		reply := "⚠️[]代表空格\n记录流水账：\n" +
			"输入：!事件[]金额\n" +
			"如：!陈先生新华都买菜 150\n\n" +
			"查看流水账：\n" +
			"输入：查看\n" +
			"如：查看\n\n" +
			"编辑流水账：\n" +
			"输入：编辑[]编号[]事件[]金额\n"+
			"如：编辑 1 买拖把 160\n\n" +
			"删除流水账：\n" +
			"输入：删除[]编号[]编号[]编号...\n" +
			"如：删除 1 2 3 4"

		return map[string]interface{}{
			"reply": reply,
		}
	} else {
		return map[string]interface{}{
			"stop": true,
		}
	}
}

// 记录流水账
func addBill(event, consumption string) (result map[string]interface{}) {
	status := true
	timeNow := time.Now()
	reply := "记录流水账失败"
	result = make(map[string]interface{})

	stmt, err := dbutil.Db.Prepare(model.InsertBillSql)
	defer stmt.Close()
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}
	_, err = stmt.Exec(event, consumption, status, timeNow, timeNow)
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}
	reply = "记录流水账成功"
	result["reply"] = reply
	return
}

// 查看流水账列表
func getBills() (result map[string]interface{}) {
	reply := "查询流水账失败"
	result = make(map[string]interface{})

	rows, err := dbutil.Db.Query(model.GetBillSqls)
	defer rows.Close()
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}

	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	reply = ""
	for rows.Next() {
		rows.Scan(scanArgs...)
		record := make(map[string]string)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			}
		}

		reply += "\n编号：" + record["id"] +
			"\n事件：" + record["event"] +
			"\n金额：" + record["consumption"] +
			"\n日期：" + record["sysDate"] + "\n"
	}

	if reply == "" {
		return map[string]interface{}{
			"reply": "暂无记录",
		}
	}
	return map[string]interface{}{
		"reply": reply,
	}
}

// 删除某条流水账
func deleteBill(strIds string) (result map[string]interface{}) {
	ids := strings.Split(strIds, " ")
	reply := "删除编号为%s的流水账失败"

	for i := 1; i < len(ids); i++ {
		result = make(map[string]interface{})

		stmt, err := dbutil.Db.Prepare(model.DeleteBillSql)
		if err != nil {
			glog.Infoln(err)
			reply = fmt.Sprintf(reply, ids[i])
			result["reply"] = reply
			return
		}
		_, err = stmt.Exec(ids[i])
		if err != nil {
			glog.Infoln(err)
			reply = fmt.Sprintf(reply, ids[i])
			result["reply"] = reply
			return
		}
	}

	reply = fmt.Sprintf("删除编号为%s的流水账成功", strIds)
	result["reply"] = reply
	return
}

// 编辑某条流水账
func updateBill(id, event, consumption string) (result map[string]interface{}) {
	reply := "编辑编号为%s的流水账失败"
	result = make(map[string]interface{})

	stmt, err := dbutil.Db.Prepare(model.UpdateBillSql)
	if err != nil {
		glog.Infoln(err)
		reply = fmt.Sprintf(reply, id)
		result["reply"] = reply
		return
	}
	_, err = stmt.Exec(event, consumption, id)
	if err != nil {
		glog.Infoln(err)
		reply = fmt.Sprintf(reply, id)
		result["reply"] = reply
		return
	}

	reply = fmt.Sprintf("编辑编号为%s的流水账成功", id)
	result["reply"] = reply
	return
}