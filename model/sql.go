package model

const (
	InsertBillSql = `
		INSERT INTO
			t_bill (event, consumption, sysDate, updateDate, uid) 
		VALUES 
			(?, ?, ?, ?, ?)
	`
	GetBillsSql = `
		SELECT 
			t_bill.id, event, consumption, t_user.name,
			date_format(sysDate, '%Y-%m-%d') as sysDate
		FROM 
			t_bill 
		LEFT JOIN
			t_user
		ON
			t_user.uid = t_bill.uid
		WHERE
			t_bill.sysDate >= ? AND t_bill.sysDate < ?
		ORDER BY
			t_bill.sysDate DESC
	`
	DeleteBillSql = `
		DELETE FROM
			t_bill 
		WHERE 
			id = ?
	`

	UpdateBillSql = `
		UPDATE
			t_bill
		SET
			event = ?, consumption = ?, updateDate = ?%s
		WHERE 
			id = ?
	`

	InsertUserSql = `
		INSERT INTO
			t_user (uid, name) 
		VALUES 
			(?, ?)
	`

	DeleteUserSql = `
		DELETE FROM
			t_user 
		WHERE 
			uid = ?
	`

	InsertBankSql = `
		INSERT INTO
			t_bank (uid, money, sysDate) 
		VALUES 
			(?, ?, ?)
	`

	CountBalanceSql = `
		SELECT
			SUM(money)
		FROM
			t_bank
	`

	CountBillsSql = `
		SELECT
			SUM(consumption)
		FROM
			t_bill
	`

	CountSpendSql = `
		SELECT
			SUM(consumption) AS consumption,
			t_user.name
		FROM
			t_bill
		LEFT JOIN
			t_user
		ON
			t_user.uid = t_bill.uid
		WHERE
			t_bill.sysDate >= ? AND t_bill.sysDate < ? 
		GROUP BY
			t_user.uid
	`

	GetBankSql = `
		SELECT
			SUM(money) as money,
			date_format(sysDate, '%Y-%m') as sDate
		FROM
			t_bank
		GROUP BY
			sDate
	`
)
