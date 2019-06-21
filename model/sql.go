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
			event = ?, consumption = ?
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
)
