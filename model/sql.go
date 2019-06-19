package model

const (
	InsertBillSql = `
		INSERT INTO
			t_bill (event, consumption, status, sysDate, updateDate) 
		VALUES 
			(?, ?, ?, ?, ?)
	`
	GetBillSqls = `
		SELECT 
			id, event, consumption,
			date_format(sysDate, '%Y-%m-%d') as sysDate
		FROM 
			t_bill 
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
)