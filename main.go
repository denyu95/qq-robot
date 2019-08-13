package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"sort"

	"github.com/denyu95/qq-robot/dbutil"
	"github.com/denyu95/qq-robot/model"
	"github.com/golang/glog"
	"github.com/juzi5201314/cqhttp-go-sdk/server"
	"github.com/360EntSecGroup-Skylar/excelize"
	"gitlab.oifitech.com/SOP/letsgo/utils"
)

func main() {
	s := server.StartListenServer(8080, "/")
	s.ListenGroupMessage(server.GroupMessageListener(group))
	s.ListenPrivateMessage(server.PrivateMessageListener(private))
	s.Listen()
}

func private(msg1 string, msg2 float64, msg3 float64, msg4 string, msg5 float64) map[string]interface{} {
	return map[string]interface{}{
		"reply": "http://www.cylovefcy.com:9999/" + export(),
	}
}

func group(m map[string]interface{}) map[string]interface{} {
	helpExp := regexp.MustCompile(`^帮助$`)

	getBillsExp := regexp.MustCompile(`^查询(?:(?:，|,)(\d{4}-\d{2}))?(?:(?:，|,)(\d{4}-\d{2}))?$`)
	deleteBillExp := regexp.MustCompile(`^删除((?:(?:,|，)(?:\d+))+)$`)
	addBillExp := regexp.MustCompile(`^(?:！|!)([^\n]+)(?:,|，)(\d+\.?\d{0,2})`)
	updateBillExp := regexp.MustCompile(`^编辑(?:,|，)(\d+)(?:,|，)([^\n]+)(?:,|，)(\d+\.?\d{0,2})(?:(?:,|，)(\d{4}-\d{2}-\d{2}))?$`)

	depositExp := regexp.MustCompile(`^充值(?:,|，)(-?\d+\.?\d{0,2})`)
	balanceExp := regexp.MustCompile(`^余额$`)
	spendExp := regexp.MustCompile(`^消费(?:(?:，|,)(\d{4}-\d{2}))?(?:(?:，|,)(\d{4}-\d{2}))?$`)

	addUserExp := regexp.MustCompile(`^用户(?:,|，)([^\n]+)(?:,|，)([^\n]+)`)

	exportExp := regexp.MustCompile(`^导出$`)

	msg := m["message"].(string)
	byteMsg := []byte(msg)

	if addBillExp.Match(byteMsg) {
		result := addBillExp.FindAllStringSubmatch(msg, -1)
		event := result[0][1]
		consumption := result[0][2]
		uid := strconv.FormatFloat(m["user_id"].(float64), 'f', -1, 64)
		return addBill(event, consumption, uid)

	} else if getBillsExp.Match(byteMsg) {
		result := getBillsExp.FindAllStringSubmatch(msg, -1)
		startDate := result[0][1]
		endDate := result[0][2]
		return getBills(startDate, endDate)

	} else if deleteBillExp.Match(byteMsg) {
		result := deleteBillExp.FindAllStringSubmatch(msg, -1)
		ids := result[0][1]
		return deleteBill(ids)

	} else if updateBillExp.Match(byteMsg) {
		result := updateBillExp.FindAllStringSubmatch(msg, -1)
		id := result[0][1]
		event := result[0][2]
		consumption := result[0][3]
		date := result[0][4]
		return updateBill(id, event, consumption, date)

	} else if helpExp.Match(byteMsg) {
		reply := "\n记录流水账：\n" +
			"输入：!事件，金额\n" +
			"如：!买菜，150\n\n" +
			"编辑流水账：\n" +
			"输入：编辑，编号，事件，金额\n" +
			"如：编辑，1，买拖把，160\n" +
			"如：编辑，1，买拖把，160，2018-01-01\n\n" +
			"删除流水账：\n" +
			"输入：删除，编号，编号，编号...\n" +
			"如：删除，1，2，3，4\n\n" +
			"充值：\n" +
			"输入：充值，金额\n" +
			"如：充值，500\n\n" +
			"查询流水账：\n" +
			"输入：查询\n\n" +
			"查询余额：\n" +
			"输入：余额\n\n" +
			"查询花费：\n" +
			"输入：花费"
		return map[string]interface{}{
			"reply": reply,
		}

	} else if spendExp.Match(byteMsg) {
		result := spendExp.FindAllStringSubmatch(msg, -1)
		startDate := result[0][1]
		endDate := result[0][2]
		return spend(startDate, endDate)

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

	} else if exportExp.Match(byteMsg) {
		return map[string]interface{}{
			"reply": "http://www.cylovefcy.com:9999/" + export(),
		}

	}else {
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

// 查询流水账列表
func getBills(startDate, endDate string) (result map[string]interface{}) {
	reply := "查询本月流水账失败"
	result = make(map[string]interface{})
	timeNow := time.Now()

	monthFirstDay := ""
	nextMonthFirstDay := ""
	if startDate != "" && endDate != "" {
		loc, _ := time.LoadLocation("Local")
		sTime, err := time.ParseInLocation("2006-01", startDate, loc)
		if err != nil {
			result["reply"] = "开始日期填写错误！"
			return
		}
		eTime, err := time.ParseInLocation("2006-01", endDate, loc)
		if err != nil {
			result["reply"] = "结束日期填写错误！"
			return
		}
		if sTime.After(eTime) {
			result["reply"] = "开始日期大于结束日期！"
			return
		}

		monthFirstDay = startDate + "-01 00:00:00"
		nextMonthFirstDay = endDate + "-01 00:00:00"
	} else if startDate != "" {
		loc, _ := time.LoadLocation("Local")
		sTime, err := time.ParseInLocation("2006-01", startDate, loc)
		if err != nil {
			result["reply"] = "开始日期填写错误！"
			return
		}
		monthFirstDay = startDate + "-01 00:00:00"
		nextMonthFirstDay = sTime.AddDate(0, 1, 0).Format("2006-01") + "-01 00:00:00"
	} else {
		monthFirstDay = timeNow.Format("2006-01") + "-01 00:00:00"
		nextMonthFirstDay = timeNow.AddDate(0, 1, 0).Format("2006-01") + "-01 00:00:00"
	}
	fmt.Println("查询 大于等于 " + monthFirstDay + " ，小于 " + nextMonthFirstDay + " 的流水账。")

	rows, err := dbutil.Db.Query(model.GetBillsSql, monthFirstDay, nextMonthFirstDay)
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

	reply = "\n流水账记录："
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

	if reply == "\n流水账记录：" {
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
func updateBill(id, event, consumption, date string) (result map[string]interface{}) {
	reply := "编辑编号为%s的流水账失败"
	result = make(map[string]interface{})

	var parameters []interface{}
	parameters = append(parameters, event, consumption, time.Now())
	strCondition := ""

	if date != "" {
		strCondition = ", sysDate = ?"
		loc, _ := time.LoadLocation("Local")
		_, err := time.ParseInLocation("2006-01-02 15:04:05", date + " 00:00:00", loc)
		if err != nil {
			result["reply"] = "日期填写错误！"
			return
		}
		parameters = append(parameters, date + " 00:00:00")
	}
	parameters = append(parameters, id)

	sql := fmt.Sprintf(model.UpdateBillSql, strCondition)

	stmt, err := dbutil.Db.Prepare(sql)
	if err != nil {
		glog.Infoln(err)
		reply = fmt.Sprintf(reply, id)
		result["reply"] = reply
		return
	}
	_, err = stmt.Exec(parameters...)
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

// 消费
func spend(startDate, endDate string) (result map[string]interface{}) {
	timeNow := time.Now()
	result = make(map[string]interface{})
	reply := "查询消费失败"

	monthFirstDay := ""
	nextMonthFirstDay := ""
	if startDate != "" && endDate != "" {
		loc, _ := time.LoadLocation("Local")
		sTime, err := time.ParseInLocation("2006-01", startDate, loc)
		if err != nil {
			result["reply"] = "开始日期填写错误！"
			return
		}
		eTime, err := time.ParseInLocation("2006-01", endDate, loc)
		if err != nil {
			result["reply"] = "结束日期填写错误！"
			return
		}
		if sTime.After(eTime) {
			result["reply"] = "开始日期大于结束日期！"
			return
		}

		monthFirstDay = startDate + "-01 00:00:00"
		nextMonthFirstDay = endDate + "-01 00:00:00"
	} else if startDate != "" {
		loc, _ := time.LoadLocation("Local")
		sTime, err := time.ParseInLocation("2006-01", startDate, loc)
		if err != nil {
			result["reply"] = "开始日期填写错误！"
			return
		}
		monthFirstDay = startDate + "-01 00:00:00"
		nextMonthFirstDay = sTime.AddDate(0, 1, 0).Format("2006-01") + "-01 00:00:00"
	} else {
		monthFirstDay = timeNow.Format("2006-01") + "-01 00:00:00"
		nextMonthFirstDay = timeNow.AddDate(0, 1, 0).Format("2006-01") + "-01 00:00:00"
	}

	fmt.Println("统计" + monthFirstDay + "-" + nextMonthFirstDay + "区间的消费。")

	rows, err := dbutil.Db.Query(model.CountSpendSql, monthFirstDay, nextMonthFirstDay)
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

		if record["sex"] == "0" {
			reply += "\n绝代佳人 " + record["name"] +
				" 消费了：" + consumption + "元！\n"
		} else {
			reply += "\n任劳任怨 " + record["name"] +
				" 消费了：" + consumption + "元！\n"
		}
	}

	if reply == "" {
		return map[string]interface{}{
			"reply": "暂无消费",
		}
	} else {
		reply += "\n总计：" + fmt.Sprintf("%.2f" ,total)
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

		return fmt.Sprintf("%.2f", i.(float64))
	case float32:
		return fmt.Sprintf("%.2f", i.(float32))
	case []byte:
		return string(i.([]byte))
	default:
		return ""
	}
}

func export() string {
	rows, err := dbutil.Db.Query(model.GetBankSql)
	defer rows.Close()
	if err != nil {
		return ""
	}
	co, _ := rows.Columns()
	sArgs := make([]interface{}, len(co))
	val := make([]interface{}, len(co))
	recs := make([]map[string]string, 0)
	a := make(map[string]string)
	allBank := 0.0
	allSpend := 0.0
	for i := range val {
		sArgs[i] = &val[i]
	}
	for rows.Next() {
		rows.Scan(sArgs...)
		record := make(map[string]string)
		for i, col := range val {
			if col != nil {
				record[co[i]] = convertString(col)
			}
		}
		recs = append(recs, record)
	}

	for _, v := range recs {
		a[v["sDate"]] = v["money"]
		floatMoney, _ := strconv.ParseFloat(v["money"], 32)
		allBank += floatMoney
	}

	rows, err = dbutil.Db.Query(model.GetBillsSql, "2000-01-01 00:00:00", "2200-01-01 00:00:00")
	defer rows.Close()
	if err != nil {
		glog.Infoln(err)
		return ""
	}

	columns, _ := rows.Columns()

	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	records := make([]map[string]string, 0)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		rows.Scan(scanArgs...)
		record := make(map[string]string)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = convertString(col)
			}
		}
		records = append(records, record)
	}

	results := make(map[string][]map[string]string, 10)
	for _, record := range records {
		sysDate := record["sysDate"][0:7]
		content := results[sysDate]
		content = append(content, record)
		results[sysDate] = content

		floatConsumption, _ := strconv.ParseFloat(record["consumption"], 32)
		allSpend += floatConsumption
	}

	f := excelize.NewFile()

	var newMp = make([]string, 0)
	for k := range results {
		newMp = append(newMp, k)
	}
	//sort.Strings(newMp)
	sort.Sort(sort.Reverse(sort.StringSlice(newMp)))
	xyz := 0.0
	for _, v := range newMp {
		las := strings.Split(v, "-")
		v3 := las[0] + "年"+ las[1] + "月"
		monthTotalSpend := 0.0
		totalRemain := allBank - allSpend - xyz
		for i, v1 := range results[v] {
			if i == 0 {
				f.NewSheet(v3)
				f.SetColWidth(v3, "A", "D", 15)
				style, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1},"alignment":{"horizontal":"center"}}`)
				if err != nil {
					fmt.Println(err)
				}
				err = f.SetCellStyle(v3, "A1", "D1", style)
				f.SetRowHeight(v3, 1, 20)
				f.SetCellValue(v3, "A1", "姓名")
				f.SetCellValue(v3, "B1", "事件")
				f.SetCellValue(v3, "C1", "金额")
				f.SetCellValue(v3, "D1", "时间")

				err = f.SetCellStyle(v3, "F1", "I1", style)
				f.SetCellValue(v3, "F1", "充值")
				f.SetCellValue(v3, "G1", "当月消费")
				f.SetCellValue(v3, "H1", "当月剩余")
				f.SetCellValue(v3, "I1", "总剩余")
			}
			f.SetRowHeight(v3, i + 2, 25)
			f.SetCellValue(v3, "A" + strconv.Itoa(i + 2), v1["name"])
			f.SetCellValue(v3, "B" + strconv.Itoa(i + 2), v1["event"])
			f.SetCellValue(v3, "C" + strconv.Itoa(i + 2), v1["consumption"])
			f.SetCellValue(v3, "D" + strconv.Itoa(i + 2), v1["sysDate"])

			consumption, _ := strconv.ParseFloat(v1["consumption"], 32)
			monthTotalSpend += consumption
		}
		bank, _ := strconv.ParseFloat(a[v], 32)

		f.SetCellValue(v3, "F2", fmt.Sprintf("%.2f", bank))
		f.SetCellValue(v3, "G2", fmt.Sprintf("%.2f", monthTotalSpend))
		f.SetCellValue(v3, "H2", fmt.Sprintf("%.2f", bank - monthTotalSpend))
		f.SetCellValue(v3, "I2", fmt.Sprintf("%.2f", totalRemain))
		xyz += bank - monthTotalSpend
	}

	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(2)

	//err = f.SaveAs("./record.xlsx")

	fileName := "record-" + utils.Uuid() + ".xlsx"

	err = f.SaveAs("/data/file/" + fileName)
	if err != nil {
		fmt.Println(err)
	}

	return  fileName
}
