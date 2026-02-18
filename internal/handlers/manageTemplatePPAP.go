package handlers

import (
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysPPAPDetail represents a row in mst_ppap_detail
type SysPPAPDetail struct {
	ID                int64            `db:"mpd_id" json:"mpd_id"`
	MmmID             int64            `db:"mmm_id" json:"mmm_id"`
	MmmModel          utils.NullString `db:"mmm_model" json:"mmm_model"`
	MmmCustomerName   utils.NullString `db:"mmm_customer_name" json:"mmm_customer_name"`
	MmmStatus         string           `db:"mmm_status" json:"mmm_status"`
	MpiName           utils.NullString `db:"mpi_name" json:"mpi_name"`
	MpiStatus         string           `db:"mpi_status" json:"mpi_status"`
	MpiID             int64            `db:"mpi_id" json:"mpi_id"`
	CreatedAt         *time.Time       `db:"mpd_created_at" json:"mpd_created_at"`
	CreatedBy         utils.NullString `db:"mpd_created_by" json:"mpd_created_by"`
	UpdatedAt         *time.Time       `db:"mpd_updated_at" json:"mpd_updated_at"`
	UpdatedBy         utils.NullString `db:"mpd_updated_by" json:"mpd_updated_by"`
	UpdateByFirstName utils.NullString `db:"mpd_updated_by_firstname" json:"mpd_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"mpd_updated_by_lastname" json:"mpd_updated_by_lastname"`
}

func ListPPAPDetails(c *fiber.Ctx, db *sqlx.DB) error {
	mmmID := c.Query("mmm_id")
	mpiID := c.Query("mpi_id")
	query := `	SELECT
				  mmm.mmm_model AS mmm_model,
				  mmm.mmm_customer_name AS mmm_customer_name,
				  mmm.mmm_status AS mmm_status,
				  mpi.mpi_name AS mpi_name,
				  mpi.mpi_status AS mpi_status,
				  mpd_id AS mpd_id,
				  mmm.mmm_id AS mmm_id,
				  mpi.mpi_id AS mpi_id,
				  mpd_updated_at AS mpd_updated_at,
				  mpd_updated_by AS mpd_updated_by,
				  su.su_firstname AS mpd_updated_by_firstname,
				  su.su_lastname AS mpd_updated_by_lastname
				FROM
				  mst_ppap_detail
				  LEFT JOIN sys_user su ON mpd_updated_by = su_emp_code
				  LEFT JOIN mst_model_master mmm ON mmm.mmm_id = mst_ppap_detail.mmm_id
				  LEFT JOIN mst_ppap_item mpi ON mpi.mpi_id = mst_ppap_detail.mpi_id
			  	WHERE 1=1`
	args := []interface{}{}
	if mmmID != "" {
		query += " AND mmm_id = ?"
		args = append(args, mmmID)
	}
	if mpiID != "" {
		query += " AND mpi_id = ?"
		args = append(args, mpiID)
	}
	query += " ORDER BY mpd_id ASC"

	var list []SysPPAPDetail
	if err := db.Select(&list, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func InsertPPAPDetail(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		MmmModel        string `json:"mmm_model"`
		MmmCustomerName string `json:"mmm_customer_name"`
		MpiID           int64  `json:"mpi_id"`
		CreatedBy       string `json:"mpd_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	// check duplicate model master entry
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_model_master WHERE mmm_model = ? AND mmm_customer_name = ?`, body.MmmModel, body.MmmCustomerName); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	res, err := db.Exec(`INSERT INTO mst_model_master (mmm_model, mmm_customer_name, mmm_status, mmm_created_at, mmm_created_by, mmm_updated_at, mmm_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?)`, body.MmmModel, body.MmmCustomerName, "active", now, body.CreatedBy, now, body.CreatedBy)
	if err != nil {
		return c.Status(500).JSON(5)
	}

	// get the last inserted id from mst_model_master and use it for mst_ppap_detail
	lastID, err := res.LastInsertId()
	if err != nil {
		return c.Status(500).JSON(5)
	}

	if _, err := db.Exec(`INSERT INTO mst_ppap_detail (mmm_id, mpi_id, mpd_created_at, mpd_created_by, mpd_updated_at, mpd_updated_by) VALUES (?, ?, ?, ?, ?, ?)`, lastID, body.MpiID, now, body.CreatedBy, now, body.CreatedBy); err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(201).JSON(1)
}

func UpdatePPAPDetail(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mpd_id"`
		MmmID     int64  `json:"mmm_id"`
		MpiID     int64  `json:"mpi_id"`
		UpdatedBy string `json:"mpd_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_ppap_detail WHERE mmm_id = ? AND mpi_id = ? AND mpd_id <> ?`, body.MmmID, body.MpiID, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE mst_ppap_detail SET mmm_id = ?, mpi_id = ?, mpd_updated_at = ?, mpd_updated_by = ? WHERE mpd_id = ?`, body.MmmID, body.MpiID, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "ppap detail not found"})
	}
	return c.Status(200).JSON(1)
}
