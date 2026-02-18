package handlers

import (
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type SysModelMaster struct {
	ID                 int64            `db:"mmm_id" json:"mmm_id"`
	Model              utils.NullString `db:"mmm_model" json:"mmm_model"`
	MpiName            utils.NullString `db:"mpi_name" json:"mpi_name"`
	Aaname             utils.NullString `db:"aname" json:"aname"`
	CustomerName       utils.NullString `db:"mmm_customer_name" json:"mmm_customer_name"`
	Status             string           `db:"mmm_status" json:"mmm_status"`
	CreatedAt          *time.Time       `db:"mmm_created_at" json:"mmm_created_at"`
	CreatedBy          utils.NullString `db:"mmm_created_by" json:"mmm_created_by"`
	UpdatedAt          *time.Time       `db:"mmm_updated_at" json:"mmm_updated_at"`
	UpdatedBy          utils.NullString `db:"mmm_updated_by" json:"mmm_updated_by"`
	UpdatedByFirstName utils.NullString `db:"mmm_updated_by_firstname" json:"mmm_updated_by_firstname"`
	UpdatedByLastName  utils.NullString `db:"mmm_updated_by_lastname" json:"mmm_updated_by_lastname"`
}

func ListModelMaster(c *fiber.Ctx, db *sqlx.DB) error {
	status := c.Query("mmm_status")
	query := `SELECT mmm.mmm_id AS mmm_id,
				 mmm_model AS mmm_model,
				 mmm_customer_name AS mmm_customer_name,
				 GROUP_CONCAT(mpi.mpi_name SEPARATOR ', ') AS aname,
				 mpi.mpi_name AS mpi_name,
				 mmm_status AS mmm_status,
				 mmm_updated_at AS mmm_updated_at,
				 mmm_updated_by AS mmm_updated_by,
				 su.su_firstname AS mmm_updated_by_firstname,
				 su.su_lastname AS mmm_updated_by_lastname
			FROM mst_model_master mmm
			LEFT JOIN sys_user su ON mmm_updated_by = su_emp_code
			LEFT JOIN mst_ppap_detail mpd ON mmm.mmm_id = mpd.mmm_id
			LEFT JOIN mst_ppap_item mpi ON mpd.mpi_id = mpi.mpi_id
			WHERE 1=1
			`
	args := []interface{}{}
	if status != "" {
		query += " AND mmm_status = ?"
		args = append(args, status)
	}
	query += " AND mpd.mpd_status = 'active' GROUP BY mmm.mmm_model,mmm.mmm_customer_name ORDER BY mmm.mmm_id ASC"

	var list []SysModelMaster
	if err := db.Select(&list, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func GetModelMaster(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var m SysModelMaster
	query := `SELECT mmm_id AS mmm_id, mmm_model AS mmm_model, mmm_customer_name AS mmm_customer_name, mmm_status AS mmm_status, mmm_created_at AS mmm_created_at, mmm_created_by AS mmm_created_by, mmm_updated_at AS mmm_updated_at, mmm_updated_by AS mmm_updated_by FROM mst_model_master WHERE mmm_id = ? LIMIT 1`
	if err := db.Get(&m, query, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(m)
}

func InsertModelMaster(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		Model        string  `json:"mmm_model"`
		CustomerName string  `json:"mmm_customer_name"`
		PpapID       []int64 `json:"mpd_id"`
		CreatedBy    string  `json:"mmm_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	// duplicate check (model + customer)
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_model_master WHERE mmm_model = ? AND mmm_customer_name = ?`, body.Model, body.CustomerName); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5)
	}

	res, err := tx.Exec(`INSERT INTO mst_model_master (mmm_model, mmm_customer_name, mmm_status, mmm_created_at, mmm_created_by, mmm_updated_at, mmm_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?)`, body.Model, body.CustomerName, "active", now, body.CreatedBy, now, body.CreatedBy)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(5)
	}
	mmmID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(5)
	}

	if len(body.PpapID) > 0 {
		stmt := `INSERT INTO mst_ppap_detail (mmm_id, mpi_id, mpd_created_at, mpd_created_by, mpd_updated_at, mpd_updated_by, mpd_status) VALUES (?, ?, ?, ?, ?, ?, 'active')`
		for _, mpiID := range body.PpapID {
			if _, err := tx.Exec(stmt, mmmID, mpiID, now, body.CreatedBy, now, body.CreatedBy); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(5)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(5)
	}

	return c.Status(201).JSON(1)
}

func UpdateModelMaster(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID           int64   `json:"mmm_id"`
		Model        string  `json:"mmm_model"`
		CustomerName string  `json:"mmm_customer_name"`
		UpdatedBy    string  `json:"mmm_updated_by"`
		PpapIDs      []int64 `json:"mpi_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_model_master WHERE mmm_model = ? AND mmm_customer_name = ? AND mmm_id <> ?`, body.Model, body.CustomerName, body.ID); err != nil {
		return c.Status(500).JSON(5.1)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5.2)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`UPDATE mst_model_master SET mmm_model = ?, mmm_customer_name = ?, mmm_updated_at = ?, mmm_updated_by = ? WHERE mmm_id = ?`, body.Model, body.CustomerName, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5.3)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "model master not found"})
	}

	// Super-update mst_ppap_detail for this model:
	// - keep any existing detail whose mpi_id is present in body.PpapIDs (set active)
	// - insert new rows for incoming mpi_ids not present
	// - mark as inactive any existing rows whose mpi_id is not in incoming list
	var existing []struct {
		MpdID  int64  `db:"mpd_id"`
		MpiID  int64  `db:"mpi_id"`
		Status string `db:"mpd_status"`
	}
	if err := tx.Select(&existing, `SELECT mpd_id, mpi_id, mpd_status FROM mst_ppap_detail WHERE mmm_id = ?`, body.ID); err != nil {
		return c.Status(500).JSON(5.4)
	}

	// build lookup of existing by mpi_id
	existingByMpi := make(map[int64]struct {
		mpdID  int64
		status string
	})
	for _, ex := range existing {
		existingByMpi[ex.MpiID] = struct {
			mpdID  int64
			status string
		}{mpdID: ex.MpdID, status: ex.Status}
	}

	// track which existing mpi_ids we've seen in the incoming list
	seen := make(map[int64]bool)

	// ensure incoming mpi_ids are present and active (insert if missing)
	for _, inMpi := range body.PpapIDs {
		if ex, ok := existingByMpi[inMpi]; ok {
			seen[inMpi] = true
			if ex.status != "active" {
				if _, err := tx.Exec(`UPDATE mst_ppap_detail SET mpd_status = 'active', mpd_updated_at = ?, mpd_updated_by = ? WHERE mpd_id = ?`, now, body.UpdatedBy, ex.mpdID); err != nil {
					return c.Status(500).JSON(5.6)
				}
			}
		} else {
			// insert new active detail
			if _, err := tx.Exec(`INSERT INTO mst_ppap_detail (mmm_id, mpi_id, mpd_created_at, mpd_created_by, mpd_updated_at, mpd_updated_by, mpd_status) VALUES (?, ?, ?, ?, ?, ?, 'active')`, body.ID, inMpi, now, body.UpdatedBy, now, body.UpdatedBy); err != nil {
				return c.Status(500).JSON(5.6)
			}
		}
	}

	// any existing mpi_id not seen should be set to inactive
	for _, ex := range existing {
		if !seen[ex.MpiID] {
			if ex.Status != "inactive" {
				if _, err := tx.Exec(`UPDATE mst_ppap_detail SET mpd_status = 'inactive', mpd_updated_at = ?, mpd_updated_by = ? WHERE mpd_id = ?`, now, body.UpdatedBy, ex.MpdID); err != nil {
					return c.Status(500).JSON(5.7)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5.8)
	}
	return c.Status(200).JSON(1)
}

func UpdateModelMasterStatus(c *fiber.Ctx, db *sqlx.DB) error {
	{
		var body struct {
			ID        int64  `json:"mmm_id"`
			Status    string `json:"mmm_status"`
			UpdatedBy string `json:"mmm_updated_by"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
		}
		if body.ID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "mmm_id is required"})
		}
		if body.Status != "active" && body.Status != "inactive" {
			return c.Status(400).JSON(fiber.Map{"error": "mmm_status must be 'active' or 'inactive'"})
		}
		now := time.Now()
		res, err := db.Exec(`UPDATE mst_model_master SET mmm_status = ?, mmm_updated_at = ?, mmm_updated_by = ? WHERE mmm_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
		if err != nil {
			return c.Status(500).JSON(5)
		}
		ra, _ := res.RowsAffected()
		if ra == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "model master not found"})
		}
		return c.Status(200).JSON(1)
	}
}

func SelectPPAPItem(c *fiber.Ctx, db *sqlx.DB) error {
	{
		var list []struct {
			ID   int64  `db:"mpi_id" json:"mpi_id"`
			Name string `db:"mpi_name" json:"mpi_name"`
		}
		query := `SELECT mpi_id, mpi_name FROM mst_ppap_item WHERE mpi_status = 'active' ORDER BY mpi_name ASC`
		if err := db.Select(&list, query); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
		}
		return c.Status(200).JSON(list)
	}
}
