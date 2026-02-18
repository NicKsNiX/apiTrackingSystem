package handlers

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"apiTrackingSystem/internal/utils"
)

// SysWorkflow represents a row in sys_workflow
type SysWorkflow struct {
	ID           int64            `db:"sw_id" json:"sw_id"`
	SdID         int64            `db:"sd_id" json:"sd_id"`
	SdDeptCode   utils.NullString `db:"sd_dept_aname" json:"sd_dept_aname"`
	SdName       utils.NullString `db:"sd_name" json:"sd_name"`
	SuID         int64            `db:"su_id" json:"su_id"`
	EmployeeCode utils.NullString `db:"su_emp_code" json:"su_emp_code"`

	Order                int64            `db:"sw_order" json:"sw_order"`
	Status               string           `db:"sw_status" json:"sw_status"`
	CreatedAt            *time.Time       `db:"sw_created_at" json:"sw_created_at"`
	CreatedBy            utils.NullString `db:"sw_created_by" json:"sw_created_by"`
	UpdatedAt            *time.Time       `db:"sw_updated_at" json:"sw_updated_at"`
	UpdatedBy            utils.NullString `db:"sw_updated_by" json:"sw_updated_by"`
	UpdateByFirstName    utils.NullString `db:"sw_updated_by_firstname" json:"sw_updated_by_firstname"`
	UpdateByLastName     utils.NullString `db:"sw_updated_by_lastname" json:"sw_updated_by_lastname"`
	SuUpdatedByFirstName utils.NullString `db:"su_updated_by_firstname" json:"su_updated_by_firstname"`
	SuUpdatedByLastName  utils.NullString `db:"su_updated_by_lastname" json:"su_updated_by_lastname"`
}

func ListWorkflow(c *fiber.Ctx, db *sqlx.DB) error {
	sdID := c.Query("sd_id")
	var res []SysWorkflow
	query := `  SELECT sw_id AS sw_id,
					 sw.sd_id AS sd_id,
					 sd.sd_dept_aname AS sd_dept_aname,
					 sd.sd_name AS sd_name,
					 sw.sw_order AS sw_order,
					 su1.su_id AS su_id,
					 su1.su_emp_code AS su_emp_code,
					 su1.su_firstname AS su_updated_by_firstname,
					 su1.su_lastname AS su_updated_by_lastname,
					 sw_status AS sw_status,
					 sw_updated_at AS sw_updated_at,
					 sw_updated_by AS sw_updated_by,
					 su.su_firstname AS sw_updated_by_firstname,
					 su.su_lastname AS sw_updated_by_lastname
				FROM sys_workflow sw
				LEFT JOIN sys_user su ON sw_updated_by = su_emp_code
				LEFT JOIN sys_user su1 ON su1.su_id = sw.su_id
				LEFT JOIN sys_department sd ON sw.sd_id = sd.sd_id
				`
	var err error
	if sdID != "" {
		query += ` WHERE sw.sd_id = ? ORDER BY sw.sw_order ASC`
		err = db.Select(&res, query, sdID)
	} else {
		query += ` ORDER BY sw.sw_order ASC`
		err = db.Select(&res, query)
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(res)
}

func InsertWorkflow(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		SdID      *int64 `json:"sd_id"`
		Order     *int64 `json:"sw_order"`
		SuID      *int64 `json:"su_id"`
		CreatedBy string `json:"sw_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	// duplicate check: same sd_id + su_id
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM sys_workflow WHERE sd_id = ?  AND su_id = ?`, body.SdID, body.SuID); err != nil {
		return c.Status(500).JSON(5)
	}

	if count > 0 {
		return c.Status(200).JSON(2)
	}

	// determine order: if provided use it, otherwise compute max(sw_order)+1 for sd_id
	var orderVal sql.NullInt64
	if body.Order != nil {
		orderVal = sql.NullInt64{Int64: *body.Order, Valid: true}
	} else {
		var maxOrder sql.NullInt64
		if err := db.Get(&maxOrder, `SELECT MAX(sw_order) FROM sys_workflow WHERE sd_id = ?`, body.SdID); err != nil {
			return c.Status(500).JSON(5)
		}
		next := int64(1)
		if maxOrder.Valid {
			next = maxOrder.Int64 + 1
		}
		orderVal = sql.NullInt64{Int64: next, Valid: true}
	}

	now := time.Now()
	if _, err := db.Exec(`INSERT INTO sys_workflow (sd_id,su_id, sw_order, sw_status, sw_created_at, sw_created_by, sw_updated_at, sw_updated_by) VALUES (?, ?, ?, 'active', ?, ?, ?, ?)`, body.SdID, body.SuID, orderVal, now, body.CreatedBy, now, body.CreatedBy); err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(201).JSON(1)
}

func UpdateWorkflow(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"sw_id"`
		SdID      *int64 `json:"sd_id"`
		SuID      *int64 `json:"su_id"`
		Order     *int64 `json:"sw_order"`
		UpdatedBy string `json:"sw_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	// duplicate check
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM sys_workflow WHERE sd_id = ? AND su_id = ? AND sw_id <> ?`, body.SdID, body.SuID, body.ID); err != nil {
		return c.Status(500).JSON(5)

	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}
	var orderVal interface{}
	if body.Order != nil {
		orderVal = *body.Order
	} else {
		orderVal = nil
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE sys_workflow SET sd_id = ?,su_id = ?, sw_order = ?, sw_updated_at = ?, sw_updated_by = ? WHERE sw_id = ?`, body.SdID, body.SuID, orderVal, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "workflow not found"})
	}
	return c.Status(200).JSON(1)
}

func UpdateWorkflowStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"sw_id"`
		Status    string `json:"sw_status"`
		UpdatedBy string `json:"sw_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "sw_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "sw_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sw_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE sys_workflow SET sw_status = ?, sw_updated_at = ?, sw_updated_by = ? WHERE sw_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "workflow not found"})
	}
	return c.Status(200).JSON(1)
}

func SelectDepartmentMW(c *fiber.Ctx, db *sqlx.DB) error {
	var departments []struct {
		ID   int64  `db:"sd_id" json:"sd_id"`
		Name string `db:"sd_name" json:"sd_name"`
		Dept string `db:"sd_dept_aname" json:"sd_dept_aname"`
	}

	query := `SELECT sd_id AS sd_id, sd_name AS sd_name, sd_dept_aname AS sd_dept_aname FROM sys_department WHERE sd_status = 'active' ORDER BY sd_name ASC`
	if err := db.Select(&departments, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(departments)
}

func SelectUserMW(c *fiber.Ctx, db *sqlx.DB) error {
	var users []struct {
		ID           int64            `db:"su_id" json:"su_id"`
		FirstName    string           `db:"su_firstname" json:"su_firstname"`
		LastName     string           `db:"su_lastname" json:"su_lastname"`
		SdID         int64            `db:"sd_id" json:"sd_id"`
		EmployeeCode utils.NullString `db:"su_emp_code" json:"su_emp_code"`
	}
	sdIDStr := c.Query("sd_id")
	if sdIDStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sd_id query parameter is required"})
	}
	sdID, err := strconv.ParseInt(sdIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "sd_id must be a number", "detail": err.Error()})
	}

	query := `SELECT su_id AS su_id, su_firstname AS su_firstname, su_lastname AS su_lastname, sd_id AS sd_id, su_emp_code AS su_emp_code FROM sys_user WHERE su_status = 'active' AND sd_id = ? ORDER BY su_firstname ASC, su_lastname ASC`
	if err := db.Select(&users, query, sdID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(users)
}
