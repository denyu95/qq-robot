package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/denyu95/qq-robot/dbutil"
	"github.com/denyu95/qq-robot/model"
	"github.com/golang/glog"
	"github.com/juzi5201314/cqhttp-go-sdk/server"
)

func main() {
	s := server.StartListenServer(8080, "/")
	s.ListenGroupMessage(server.GroupMessageListener(group))
	s.Listen()
}

func group(m map[string]interface{}) map[string]interface{} {
	helpExp := regexp.MustCompile(`^帮助$`)

	getBillsExp := regexp.MustCompile(`^查看$`)
	deleteBillExp := regexp.MustCompile(`^删除((?:(?:,|，)(?:\d+))+)$`)
	addBillExp := regexp.MustCompile(`^(?:！|!)([^\n]+)(?:,|，)(\d+\.?\d{0,2})`)
	updateBillExp := regexp.MustCompile(`^编辑(?:,|，)(\d+)(?:,|，)([^\n]+)(?:,|，)(\d+\.?\d{0,2})`)

	depositExp := regexp.MustCompile(`^充值(?:,|，)(-?\d+\.?\d{0,2})`)
	balanceExp := regexp.MustCompile(`^余额$`)
	spendExp := regexp.MustCompile(`^花费$`)

	addUserExp := regexp.MustCompile(`^用户(?:,|，)([^\n]+)(?:,|，)([^\n]+)`)

	msg := m["message"].(string)
	byteMsg := []byte(msg)

	if addBillExp.Match(byteMsg) {
		result := addBillExp.FindAllStringSubmatch(msg, -1)
		event := result[0][1]
		consumption := result[0][2]
		uid := strconv.FormatFloat(m["user_id"].(float64), 'f', -1, 64)
		return addBill(event, consumption, uid)

	} else if getBillsExp.Match(byteMsg) {
		return getBills()

	} else if deleteBillExp.Match(byteMsg) {
		result := deleteBillExp.FindAllStringSubmatch(msg, -1)
		ids := result[0][1]
		return deleteBill(ids)

	} else if updateBillExp.Match(byteMsg) {
		result := updateBillExp.FindAllStringSubmatch(msg, -1)
		id := result[0][1]
		event := result[0][2]
		consumption := result[0][3]
		return updateBill(id, event, consumption)

	} else if helpExp.Match(byteMsg) {
		reply := "\n记录流水账：\n" +
			"输入：!事件，金额\n" +
			"如：!买菜，150\n\n" +
			"编辑流水账：\n" +
			"输入：编辑，编号，事件，金额\n" +
			"如：编辑，1，买拖把，160\n\n" +
			"删除流水账：\n" +
			"输入：删除，编号，编号，编号...\n" +
			"如：删除，1，2，3，4\n\n" +
			"充值：\n" +
			"输入：充值，金额\n" +
			"如：充值，500\n\n" +
			"查看流水账：\n" +
			"输入：查看\n\n" +
			"查看余额：\n" +
			"输入：余额\n\n" +
			"查看花费：\n" +
			"输入：花费"
		return map[string]interface{}{
			"reply": reply,
		}

	} else if spendExp.Match(byteMsg) {
		return spend()

	} else if addUserExp.Match(byteMsg) {
		result := addUserExp.FindAllStringSubmatch(msg, -1)
		uid := result[0][1]
		name := result[0][2]
		return addUser(uid, name)

	} else if depositExp.Match(byteMsg) {
		result := depositExp.FindAllStringSubmatch(msg, -1)
		money := result[0][1]
		uid := strconv.FormatFloat(m["user_id"].(float64), 'f', -1, 64)
		return deposit(uid, money)

	} else if balanceExp.Match(byteMsg) {
		return balance()

	} else {
		return map[string]interface{}{
			"stop": true,
		}
	}
}

// 记录流水账
func addBill(event, consumption, uid string) (result map[string]interface{}) {
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
	_, err = stmt.Exec(event, consumption, timeNow, timeNow, uid)
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
	reply := "查询本月流水账失败"
	result = make(map[string]interface{})

	monthFirstDay := time.Now().Format("2006-01") + "-01 00:00:00"
	fmt.Println("查询大于" + monthFirstDay + "的流水账。")

	rows, err := dbutil.Db.Query(model.GetBillsSql, monthFirstDay)
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

	reply = "\n本月流水账："
	for rows.Next() {
		rows.Scan(scanArgs...)
		record := make(map[string]string)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = convertString(col)
			}
		}

		reply += "\n编号：" + record["id"] +
			"\n事件：" + record["event"] +
			"\n金额：" + record["consumption"] + "元" +
			"\n日期：" + record["sysDate"] +
			"\n记录人：" + record["name"] + "\n"
	}

	if reply == "\n本月流水账：" {
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
	ids := strings.Split(strIds, "，")
	if len(ids) < 2 {
		ids = strings.Split(strIds, ",")
	}
	reply := "删除编号为%s的流水账失败"
	result = make(map[string]interface{})

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
	_, err = stmt.Exec(event, consumption, time.Now(), id)
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

// 添加用户
func addUser(uid, name string) (result map[string]interface{}) {
	reply := "添加用户失败"
	result = make(map[string]interface{})

	stmt, err := dbutil.Db.Prepare(model.DeleteUserSql)
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}
	_, err = stmt.Exec(uid)
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}

	stmt, err = dbutil.Db.Prepare(model.InsertUserSql)
	defer stmt.Close()
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}
	_, err = stmt.Exec(uid, name)
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}
	reply = "添加用户成功"
	result["reply"] = reply
	return
}

// 充值
func deposit(uid string, money string) (result map[string]interface{}) {
	timeNow := time.Now()
	reply := "充值失败"
	result = make(map[string]interface{})

	stmt, err := dbutil.Db.Prepare(model.InsertBankSql)
	defer stmt.Close()
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}
	_, err = stmt.Exec(uid, money, timeNow)
	if err != nil {
		glog.Infoln(err)
		result["reply"] = reply
		return
	}
	reply = "充值成功"
	result["reply"] = reply
	return
}

// 余额
func balance() (result map[string]interface{}) {
	result = make(map[string]interface{})

	row := dbutil.Db.QueryRow(model.CountBalanceSql)
	var balance float32
	err := row.Scan(&balance)
	if err != nil {
		balance = 0
	}

	row = dbutil.Db.QueryRow(model.CountBillsSql)
	var consumption float32
	err = row.Scan(&consumption)
	if err != nil {
		consumption = 0
	}

	reply := "余额：%.2f元"
	result["reply"] = fmt.Sprintf(reply, balance-consumption)

	return
}

// 花费
func spend() (result map[string]interface{}) {
	result = make(map[string]interface{})
	reply := "查询花费失败"

	monthFirstDay := time.Now().Format("2006-01") + "-01 00:00:00"
	fmt.Println("统计" + monthFirstDay + "花费。")

	rows, err := dbutil.Db.Query(model.CountSpendSql, monthFirstDay)
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

	var total float64
	for rows.Next() {
		rows.Scan(scanArgs...)
		record := make(map[string]string)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = convertString(col)
			}
		}
		fmt.Println(record)
		consumption := record["consumption"]
		floatConsumption, _ := strconv.ParseFloat(consumption, 64)
		total += floatConsumption

		reply += "\n坏人：" + record["name"] +
			"，竟然消费了：" + consumption + "元！\n"
	}

	if reply == "" {
		return map[string]interface{}{
			"reply": "暂无消费",
		}
	} else {
		reply += "\n总计：" + strconv.FormatFloat(total, 'f', -1, 32)
		return map[string]interface{}{
			"reply": reply,
		}
	}
}

func convertString(i interface{}) string {
	switch i.(type) {
	case string:
		return i.(string)
	case int64:
		return strconv.FormatInt(i.(int64), 10)
	case int32:
		return strconv.Itoa(i.(int))
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(i.(float32)), 'f', -1, 32)
	case []byte:
		return string(i.([]byte))
	default:
		return ""
	}
}
