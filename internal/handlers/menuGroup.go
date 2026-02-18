package handlers

import (
	"apiTrackingSystem/internal/utils"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysMenu represents a row in sys_menu
type SysMenu struct {
	ID                int64            `db:"sm_id" json:"sm_id"`
	Name              utils.NullString `db:"sm_name" json:"sm_name"`
	Icon              utils.NullString `db:"sm_icon" json:"sm_icon"`
	Order             int64            `db:"sm_order" json:"sm_order"`
	Status            string           `db:"sm_status" json:"sm_status"`
	UpdatedAt         *time.Time       `db:"sm_updated_at" json:"sm_updated_at"`
	UpdatedBy         utils.NullString `db:"sm_updated_by" json:"updated_by"`
	UpdateByFirstName utils.NullString `db:"sm_updated_by_firstname" json:"sm_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"sm_updated_by_lastname" json:"sm_updated_by_lastname"`
}

// ListMenus returns all menu records
func ListMenusGroup(c *fiber.Ctx, db *sqlx.DB) error {
	var menus []SysMenu
	query := `SELECT 
					sm.sm_id AS sm_id,
					sm.sm_name AS sm_name,
					sm.sm_icon AS sm_icon,
					sm.sm_order AS sm_order,
					sm.sm_status AS sm_status,
					sm.sm_updated_at AS sm_updated_at,
					sm.sm_updated_by AS sm_updated_by,
					su.su_firstname AS sm_updated_by_firstname,
					su.su_lastname AS sm_updated_by_lastname
				FROM sys_menu sm 
				LEFT JOIN sys_user su ON sm.sm_updated_by = su.su_emp_code
	 			ORDER BY sm.sm_order ASC`
	if err := db.Select(&menus, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(menus)
}

// GetMenu returns a single menu by id
func GetMenuGroup(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Query("sm_id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var m SysMenu
	query := `SELECT 
					sm.sm_id AS sm_id,
					sm.sm_name AS sm_name,
					sm.sm_icon AS sm_icon,
					sm.sm_order AS sm_order,
					sm.sm_status AS sm_status,
					sm.sm_updated_at AS sm_updated_at,
					sm.sm_updated_by AS sm_updated_by 
				FROM sys_menu sm 
				WHERE sm.sm_id = ? LIMIT 1`
	if err := db.Get(&m, query, id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "menu not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(m)
}

func GetSelectMenuGroup(c *fiber.Ctx, db *sqlx.DB) error {
	var menus []SysMenu
	query := `SELECT sm.sm_id AS sm_id, sm.sm_name AS sm_name, sm_icon, sm_order FROM sys_menu sm WHERE sm.sm_status = 'active' ORDER BY sm.sm_order ASC`
	if err := db.Select(&menus, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(menus)
}

func InsertMenuGroup(c *fiber.Ctx, db *sqlx.DB) error {
	// Parse into a plain request struct to avoid BodyParser errors with custom types
	var body struct {
		Name      string `json:"sm_name"`
		Icon      string `json:"sm_icon"`
		CreatedBy string `json:"sm_created_by"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sm_name is required"})
	}
	now := time.Now()

	// compute sm_order as MAX(sm_order) + 1 (first record gets 1)
	var maxOrder sql.NullInt64
	if err := db.Get(&maxOrder, `SELECT MAX(sm_order) FROM sys_menu`); err != nil {
		return c.Status(500).JSON(5)
	}
	nextOrder := int64(1)
	if maxOrder.Valid {
		nextOrder = maxOrder.Int64 + 1
	}
	orderVal := sql.NullInt64{Int64: nextOrder, Valid: true}
	// check if sm_name already exists
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM sys_menu WHERE sm_name = ?`, body.Name); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	_, err := db.Exec(`INSERT INTO sys_menu (sm_name, sm_icon, sm_order, sm_status, sm_created_at, sm_created_by, sm_updated_at, sm_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		body.Name, body.Icon, orderVal, "active", now, body.CreatedBy, now, body.CreatedBy)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(200).JSON(1)
}

func UpdateMenuGroupStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		SmID      int64  `json:"sm_id"`
		Status    string `json:"sm_status"`
		UpdatedBy string `json:"sm_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	_, err := db.Exec(`UPDATE sys_menu SET sm_status = ?, sm_updated_at = ?, sm_updated_by = ? WHERE sm_id = ?`,
		body.Status, time.Now(), body.UpdatedBy, body.SmID)
	if err != nil {
		return c.Status(200).JSON(0)
	}
	return c.Status(200).JSON(1)
}

func UpdateMenuGroup(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"sm_id"`
		Name      string `json:"sm_name"`
		Icon      string `json:"sm_icon"`
		Order     int64  `json:"sm_order"`
		UpdatedBy string `json:"updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "sm_id is required"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sm_name is required"})
	}

	// check for duplicate sm_name (exclude current record)
	if body.Name != "" {
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM sys_menu WHERE sm_name = ? AND sm_id <> ?`, body.Name, body.ID); err != nil {
			return c.Status(500).JSON(6)
		}
		if count > 0 {
			return c.Status(200).JSON(2)
		}
	}

	_, err := db.Exec(`UPDATE sys_menu SET sm_name = ?, sm_icon = ?, sm_order = ?, sm_updated_at = ?, sm_updated_by = ? WHERE sm_id = ?`,
		body.Name, body.Icon, body.Order, time.Now(), body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(200).JSON(1)
}
