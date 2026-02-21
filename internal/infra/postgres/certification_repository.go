package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CertificationRepository struct {
	DB *pgxpool.Pool
}

// --- Certification Types ---

func (r *CertificationRepository) ListCertTypes(ctx context.Context) ([]domain.CertificationType, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, code, name, authority, has_expiry, recency_required_months, recency_period_months, created_at
		FROM certification_types
		ORDER BY authority, code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CertificationType
	for rows.Next() {
		ct, err := scanCertType(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, ct)
	}
	return items, rows.Err()
}

func (r *CertificationRepository) GetCertTypeByID(ctx context.Context, id uuid.UUID) (domain.CertificationType, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, code, name, authority, has_expiry, recency_required_months, recency_period_months, created_at
		FROM certification_types
		WHERE id=$1
	`, id)
	return scanCertType(row)
}

func scanCertType(row pgx.Row) (domain.CertificationType, error) {
	var ct domain.CertificationType
	if err := row.Scan(&ct.ID, &ct.Code, &ct.Name, &ct.Authority, &ct.HasExpiry, &ct.RecencyRequiredMonths, &ct.RecencyPeriodMonths, &ct.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.CertificationType{}, domain.ErrNotFound
		}
		return domain.CertificationType{}, err
	}
	return ct, nil
}

// --- Employee Certifications ---

func (r *CertificationRepository) GetCertByID(ctx context.Context, orgID, id uuid.UUID) (domain.EmployeeCertification, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, user_id, cert_type_id, certificate_number, issue_date, expiry_date, status, verified_by, verified_at, document_url, created_at, updated_at
		FROM employee_certifications
		WHERE org_id=$1 AND id=$2
	`, orgID, id)
	return scanEmployeeCert(row)
}

func (r *CertificationRepository) ListCertsByUser(ctx context.Context, orgID, userID uuid.UUID) ([]domain.EmployeeCertification, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, user_id, cert_type_id, certificate_number, issue_date, expiry_date, status, verified_by, verified_at, document_url, created_at, updated_at
		FROM employee_certifications
		WHERE org_id=$1 AND user_id=$2
		ORDER BY created_at DESC
	`, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.EmployeeCertification
	for rows.Next() {
		c, err := scanEmployeeCert(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *CertificationRepository) CreateCert(ctx context.Context, cert domain.EmployeeCertification) (domain.EmployeeCertification, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO employee_certifications
			(id, org_id, user_id, cert_type_id, certificate_number, issue_date, expiry_date, status, verified_by, verified_at, document_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, org_id, user_id, cert_type_id, certificate_number, issue_date, expiry_date, status, verified_by, verified_at, document_url, created_at, updated_at
	`, cert.ID, cert.OrgID, cert.UserID, cert.CertTypeID, cert.CertificateNumber, cert.IssueDate, cert.ExpiryDate, cert.Status, cert.VerifiedBy, cert.VerifiedAt, cert.DocumentURL, cert.CreatedAt, cert.UpdatedAt)
	created, err := scanEmployeeCert(row)
	if err != nil {
		return domain.EmployeeCertification{}, TranslateError(err)
	}
	return created, nil
}

func (r *CertificationRepository) UpdateCert(ctx context.Context, cert domain.EmployeeCertification) (domain.EmployeeCertification, error) {
	row := r.DB.QueryRow(ctx, `
		UPDATE employee_certifications
		SET certificate_number=$1, expiry_date=$2, status=$3, verified_by=$4, verified_at=$5, document_url=$6, updated_at=$7
		WHERE org_id=$8 AND id=$9
		RETURNING id, org_id, user_id, cert_type_id, certificate_number, issue_date, expiry_date, status, verified_by, verified_at, document_url, created_at, updated_at
	`, cert.CertificateNumber, cert.ExpiryDate, cert.Status, cert.VerifiedBy, cert.VerifiedAt, cert.DocumentURL, cert.UpdatedAt, cert.OrgID, cert.ID)
	updated, err := scanEmployeeCert(row)
	if err != nil {
		return domain.EmployeeCertification{}, TranslateError(err)
	}
	return updated, nil
}

func (r *CertificationRepository) ListExpiringCerts(ctx context.Context, orgID uuid.UUID, before time.Time) ([]domain.EmployeeCertification, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, user_id, cert_type_id, certificate_number, issue_date, expiry_date, status, verified_by, verified_at, document_url, created_at, updated_at
		FROM employee_certifications
		WHERE org_id=$1 AND status='active' AND expiry_date IS NOT NULL AND expiry_date <= $2
		ORDER BY expiry_date ASC
	`, orgID, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.EmployeeCertification
	for rows.Next() {
		c, err := scanEmployeeCert(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func scanEmployeeCert(row pgx.Row) (domain.EmployeeCertification, error) {
	var c domain.EmployeeCertification
	if err := row.Scan(&c.ID, &c.OrgID, &c.UserID, &c.CertTypeID, &c.CertificateNumber, &c.IssueDate, &c.ExpiryDate, &c.Status, &c.VerifiedBy, &c.VerifiedAt, &c.DocumentURL, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.EmployeeCertification{}, domain.ErrNotFound
		}
		return domain.EmployeeCertification{}, err
	}
	return c, nil
}

// --- Type Ratings ---

func (r *CertificationRepository) ListTypeRatingsByUser(ctx context.Context, orgID, userID uuid.UUID) ([]domain.EmployeeTypeRating, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, user_id, aircraft_type_id, cert_id, endorsement_date, status, created_at
		FROM employee_type_ratings
		WHERE org_id=$1 AND user_id=$2
		ORDER BY created_at DESC
	`, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.EmployeeTypeRating
	for rows.Next() {
		tr, err := scanTypeRating(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, tr)
	}
	return items, rows.Err()
}

func (r *CertificationRepository) CreateTypeRating(ctx context.Context, rating domain.EmployeeTypeRating) (domain.EmployeeTypeRating, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO employee_type_ratings
			(id, org_id, user_id, aircraft_type_id, cert_id, endorsement_date, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, org_id, user_id, aircraft_type_id, cert_id, endorsement_date, status, created_at
	`, rating.ID, rating.OrgID, rating.UserID, rating.AircraftTypeID, rating.CertID, rating.EndorsementDate, rating.Status, rating.CreatedAt)
	created, err := scanTypeRating(row)
	if err != nil {
		return domain.EmployeeTypeRating{}, TranslateError(err)
	}
	return created, nil
}

func (r *CertificationRepository) HasTypeRating(ctx context.Context, orgID, userID, aircraftTypeID uuid.UUID) (bool, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT 1
		FROM employee_type_ratings
		WHERE org_id=$1 AND user_id=$2 AND aircraft_type_id=$3 AND status='active'
		LIMIT 1
	`, orgID, userID, aircraftTypeID)
	var marker int
	if err := row.Scan(&marker); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func scanTypeRating(row pgx.Row) (domain.EmployeeTypeRating, error) {
	var tr domain.EmployeeTypeRating
	if err := row.Scan(&tr.ID, &tr.OrgID, &tr.UserID, &tr.AircraftTypeID, &tr.CertID, &tr.EndorsementDate, &tr.Status, &tr.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.EmployeeTypeRating{}, domain.ErrNotFound
		}
		return domain.EmployeeTypeRating{}, err
	}
	return tr, nil
}

// --- Skills ---

func (r *CertificationRepository) ListSkillsByUser(ctx context.Context, orgID, userID uuid.UUID) ([]domain.EmployeeSkill, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, user_id, skill_type_id, proficiency_level, qualification_date, expiry_date, created_at
		FROM employee_skills
		WHERE org_id=$1 AND user_id=$2
		ORDER BY created_at DESC
	`, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.EmployeeSkill
	for rows.Next() {
		s, err := scanSkill(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *CertificationRepository) CreateSkill(ctx context.Context, skill domain.EmployeeSkill) (domain.EmployeeSkill, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO employee_skills
			(id, org_id, user_id, skill_type_id, proficiency_level, qualification_date, expiry_date, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, org_id, user_id, skill_type_id, proficiency_level, qualification_date, expiry_date, created_at
	`, skill.ID, skill.OrgID, skill.UserID, skill.SkillTypeID, skill.ProficiencyLevel, skill.QualificationDate, skill.ExpiryDate, skill.CreatedAt)
	created, err := scanSkill(row)
	if err != nil {
		return domain.EmployeeSkill{}, TranslateError(err)
	}
	return created, nil
}

func (r *CertificationRepository) UpdateSkill(ctx context.Context, skill domain.EmployeeSkill) (domain.EmployeeSkill, error) {
	row := r.DB.QueryRow(ctx, `
		UPDATE employee_skills
		SET proficiency_level=$1, expiry_date=$2
		WHERE org_id=$3 AND id=$4
		RETURNING id, org_id, user_id, skill_type_id, proficiency_level, qualification_date, expiry_date, created_at
	`, skill.ProficiencyLevel, skill.ExpiryDate, skill.OrgID, skill.ID)
	updated, err := scanSkill(row)
	if err != nil {
		return domain.EmployeeSkill{}, TranslateError(err)
	}
	return updated, nil
}

func scanSkill(row pgx.Row) (domain.EmployeeSkill, error) {
	var s domain.EmployeeSkill
	if err := row.Scan(&s.ID, &s.OrgID, &s.UserID, &s.SkillTypeID, &s.ProficiencyLevel, &s.QualificationDate, &s.ExpiryDate, &s.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.EmployeeSkill{}, domain.ErrNotFound
		}
		return domain.EmployeeSkill{}, err
	}
	return s, nil
}

// --- Recency ---

func (r *CertificationRepository) LogRecency(ctx context.Context, entry domain.EmployeeRecencyLog) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO employee_recency_log
			(id, org_id, user_id, aircraft_type_id, task_id, work_date, hours_logged, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, entry.ID, entry.OrgID, entry.UserID, entry.AircraftTypeID, entry.TaskID, entry.WorkDate, entry.HoursLogged, entry.CreatedAt)
	return TranslateError(err)
}

func (r *CertificationRepository) GetRecencyHours(ctx context.Context, orgID, userID, aircraftTypeID uuid.UUID, since time.Time) (float64, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(hours_logged), 0)
		FROM employee_recency_log
		WHERE org_id=$1 AND user_id=$2 AND aircraft_type_id=$3 AND work_date >= $4
	`, orgID, userID, aircraftTypeID, since)
	var hours float64
	if err := row.Scan(&hours); err != nil {
		return 0, err
	}
	return hours, nil
}

// --- Task Skill Requirements ---

func (r *CertificationRepository) ListRequirements(ctx context.Context, orgID uuid.UUID, taskType domain.TaskType, aircraftTypeID *uuid.UUID) ([]domain.TaskSkillRequirement, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, org_id, task_type, aircraft_type_id, cert_type_id, skill_type_id, min_proficiency_level, is_certifying_role, is_inspection_role, created_at
		FROM task_skill_requirements
		WHERE org_id=$1 AND task_type=$2 AND (aircraft_type_id IS NULL OR aircraft_type_id=$3)
		ORDER BY created_at
	`, orgID, taskType, aircraftTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.TaskSkillRequirement
	for rows.Next() {
		var req domain.TaskSkillRequirement
		if err := rows.Scan(&req.ID, &req.OrgID, &req.TaskType, &req.AircraftTypeID, &req.CertTypeID, &req.SkillTypeID, &req.MinProficiencyLevel, &req.IsCertifyingRole, &req.IsInspectionRole, &req.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, req)
	}
	return items, rows.Err()
}

func (r *CertificationRepository) CreateRequirement(ctx context.Context, req domain.TaskSkillRequirement) (domain.TaskSkillRequirement, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO task_skill_requirements
			(id, org_id, task_type, aircraft_type_id, cert_type_id, skill_type_id, min_proficiency_level, is_certifying_role, is_inspection_role, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, org_id, task_type, aircraft_type_id, cert_type_id, skill_type_id, min_proficiency_level, is_certifying_role, is_inspection_role, created_at
	`, req.ID, req.OrgID, req.TaskType, req.AircraftTypeID, req.CertTypeID, req.SkillTypeID, req.MinProficiencyLevel, req.IsCertifyingRole, req.IsInspectionRole, req.CreatedAt)
	var created domain.TaskSkillRequirement
	if err := row.Scan(&created.ID, &created.OrgID, &created.TaskType, &created.AircraftTypeID, &created.CertTypeID, &created.SkillTypeID, &created.MinProficiencyLevel, &created.IsCertifyingRole, &created.IsInspectionRole, &created.CreatedAt); err != nil {
		return domain.TaskSkillRequirement{}, TranslateError(err)
	}
	return created, nil
}

// --- Qualified Mechanics ---

func (r *CertificationRepository) GetQualifiedMechanics(ctx context.Context, orgID uuid.UUID, taskType domain.TaskType, aircraftTypeID *uuid.UUID) ([]uuid.UUID, error) {
	// Get all mechanics in the org
	rows, err := r.DB.Query(ctx, `
		SELECT DISTINCT u.id
		FROM users u
		WHERE u.org_id=$1 AND u.role='mechanic' AND u.deleted_at IS NULL
		  AND NOT EXISTS (
			-- Check each required certification is held
			SELECT 1 FROM task_skill_requirements tsr
			WHERE tsr.org_id=$1 AND tsr.task_type=$2
			  AND (tsr.aircraft_type_id IS NULL OR tsr.aircraft_type_id=$3)
			  AND tsr.cert_type_id IS NOT NULL
			  AND NOT EXISTS (
				SELECT 1 FROM employee_certifications ec
				WHERE ec.org_id=$1 AND ec.user_id=u.id AND ec.cert_type_id=tsr.cert_type_id
				  AND ec.status='active'
				  AND (ec.expiry_date IS NULL OR ec.expiry_date > now())
			  )
		  )
		  AND NOT EXISTS (
			-- Check each required skill is held
			SELECT 1 FROM task_skill_requirements tsr
			WHERE tsr.org_id=$1 AND tsr.task_type=$2
			  AND (tsr.aircraft_type_id IS NULL OR tsr.aircraft_type_id=$3)
			  AND tsr.skill_type_id IS NOT NULL
			  AND NOT EXISTS (
				SELECT 1 FROM employee_skills es
				WHERE es.org_id=$1 AND es.user_id=u.id AND es.skill_type_id=tsr.skill_type_id
				  AND es.proficiency_level >= tsr.min_proficiency_level
				  AND (es.expiry_date IS NULL OR es.expiry_date > now())
			  )
		  )
		ORDER BY u.id
	`, orgID, taskType, aircraftTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
