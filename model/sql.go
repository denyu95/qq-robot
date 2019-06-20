package model

const (
	InsertBillSql = `
		INSERT INTO
			t_bill (event, consumption, status, sysDate, updateDate, uid) 
		VALUES 
			(?, ?, ?, ?, ?, ?)
	`
	GetBillSqls = `
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
			status = 1
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
)
