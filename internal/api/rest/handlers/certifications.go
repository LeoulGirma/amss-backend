package handlers

import (
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// --- Request/Response Types ---

type certCreateRequest struct {
	CertTypeID        string `json:"cert_type_id" validate:"required,uuid"`
	CertificateNumber string `json:"certificate_number" validate:"required"`
	IssueDate         string `json:"issue_date" validate:"required,rfc3339"`
	ExpiryDate        string `json:"expiry_date" validate:"omitempty,rfc3339"`
	DocumentURL       string `json:"document_url" validate:"omitempty"`
}

type certUpdateRequest struct {
	Status      *string `json:"status" validate:"omitempty,oneof=active expired suspended revoked"`
	ExpiryDate  *string `json:"expiry_date" validate:"omitempty,rfc3339"`
	DocumentURL *string `json:"document_url" validate:"omitempty"`
}

type typeRatingCreateRequest struct {
	AircraftTypeID  string `json:"aircraft_type_id" validate:"required,uuid"`
	CertID          string `json:"cert_id" validate:"required,uuid"`
	EndorsementDate string `json:"endorsement_date" validate:"required,rfc3339"`
}

type skillCreateRequest struct {
	SkillTypeID       string `json:"skill_type_id" validate:"required,uuid"`
	ProficiencyLevel  int    `json:"proficiency_level" validate:"min=1,max=5"`
	QualificationDate string `json:"qualification_date" validate:"required,rfc3339"`
	ExpiryDate        string `json:"expiry_date" validate:"omitempty,rfc3339"`
}

type requirementCreateRequest struct {
	TaskType            string  `json:"task_type" validate:"required,oneof=inspection repair overhaul"`
	AircraftTypeID      *string `json:"aircraft_type_id" validate:"omitempty,uuid"`
	CertTypeID          *string `json:"cert_type_id" validate:"omitempty,uuid"`
	SkillTypeID         *string `json:"skill_type_id" validate:"omitempty,uuid"`
	MinProficiencyLevel int     `json:"min_proficiency_level" validate:"min=0,max=5"`
	IsCertifyingRole    bool    `json:"is_certifying_role"`
	IsInspectionRole    bool    `json:"is_inspection_role"`
}

type certResponse struct {
	ID                uuid.UUID                  `json:"id"`
	OrgID             uuid.UUID                  `json:"org_id"`
	UserID            uuid.UUID                  `json:"user_id"`
	CertTypeID        uuid.UUID                  `json:"cert_type_id"`
	CertificateNumber string                     `json:"certificate_number"`
	IssueDate         time.Time                  `json:"issue_date"`
	ExpiryDate        *time.Time                 `json:"expiry_date,omitempty"`
	Status            domain.CertificationStatus `json:"status"`
	VerifiedBy        *uuid.UUID                 `json:"verified_by,omitempty"`
	VerifiedAt        *time.Time                 `json:"verified_at,omitempty"`
	DocumentURL       string                     `json:"document_url,omitempty"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

type typeRatingResponse struct {
	ID              uuid.UUID                `json:"id"`
	OrgID           uuid.UUID                `json:"org_id"`
	UserID          uuid.UUID                `json:"user_id"`
	AircraftTypeID  uuid.UUID                `json:"aircraft_type_id"`
	CertID          uuid.UUID                `json:"cert_id"`
	EndorsementDate time.Time                `json:"endorsement_date"`
	Status          domain.TypeRatingStatus  `json:"status"`
	CreatedAt       time.Time                `json:"created_at"`
}

type skillResponse struct {
	ID                uuid.UUID  `json:"id"`
	OrgID             uuid.UUID  `json:"org_id"`
	UserID            uuid.UUID  `json:"user_id"`
	SkillTypeID       uuid.UUID  `json:"skill_type_id"`
	ProficiencyLevel  int        `json:"proficiency_level"`
	QualificationDate time.Time  `json:"qualification_date"`
	ExpiryDate        *time.Time `json:"expiry_date,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type certTypeResponse struct {
	ID                    uuid.UUID                     `json:"id"`
	Code                  string                        `json:"code"`
	Name                  string                        `json:"name"`
	Authority             domain.CertificationAuthority `json:"authority"`
	HasExpiry             bool                          `json:"has_expiry"`
	RecencyRequiredMonths *int                          `json:"recency_required_months,omitempty"`
	RecencyPeriodMonths   *int                          `json:"recency_period_months,omitempty"`
}

type qualificationCheckResponse struct {
	Qualified           bool     `json:"qualified"`
	MissingCerts        []string `json:"missing_certs,omitempty"`
	MissingTypeRatings  []string `json:"missing_type_ratings,omitempty"`
	MissingSkills       []string `json:"missing_skills,omitempty"`
	ExpiredCerts        []string `json:"expired_certs,omitempty"`
	InsufficientRecency bool     `json:"insufficient_recency,omitempty"`
	Reasons             []string `json:"reasons,omitempty"`
}

// --- Handlers ---

func ListCertTypes(w http.ResponseWriter, r *http.Request) {
	_, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	types, err := servicesReg.Certifications.ListCertTypes(r.Context())
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]certTypeResponse, 0, len(types))
	for _, ct := range types {
		resp = append(resp, certTypeResponse{
			ID:                    ct.ID,
			Code:                  ct.Code,
			Name:                  ct.Name,
			Authority:             ct.Authority,
			HasExpiry:             ct.HasExpiry,
			RecencyRequiredMonths: ct.RecencyRequiredMonths,
			RecencyPeriodMonths:   ct.RecencyPeriodMonths,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func ListUserCertifications(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}
	certs, err := servicesReg.Certifications.ListCertsByUser(r.Context(), actor, userID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]certResponse, 0, len(certs))
	for _, c := range certs {
		resp = append(resp, mapCert(c))
	}
	writeJSON(w, http.StatusOK, resp)
}

func CreateUserCertification(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}
	var req certCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	certTypeID, _ := uuid.Parse(req.CertTypeID)
	issueDate, _ := time.Parse(time.RFC3339, req.IssueDate)
	var expiryDate *time.Time
	if req.ExpiryDate != "" {
		v, _ := time.Parse(time.RFC3339, req.ExpiryDate)
		expiryDate = &v
	}

	created, err := servicesReg.Certifications.CreateCert(r.Context(), actor, services.CertCreateInput{
		UserID:            userID,
		CertTypeID:        certTypeID,
		CertificateNumber: req.CertificateNumber,
		IssueDate:         issueDate,
		ExpiryDate:        expiryDate,
		DocumentURL:       req.DocumentURL,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapCert(created))
}

func UpdateUserCertification(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	certID, err := uuid.Parse(chi.URLParam(r, "certId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid certification id")
		return
	}
	var req certUpdateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	var status *domain.CertificationStatus
	if req.Status != nil {
		v := domain.CertificationStatus(*req.Status)
		status = &v
	}
	var expiryDate *time.Time
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		v, _ := time.Parse(time.RFC3339, *req.ExpiryDate)
		expiryDate = &v
	}

	updated, err := servicesReg.Certifications.UpdateCert(r.Context(), actor, certID, status, expiryDate, req.DocumentURL)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapCert(updated))
}

func ListUserTypeRatings(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}
	ratings, err := servicesReg.Certifications.ListTypeRatingsByUser(r.Context(), actor, userID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]typeRatingResponse, 0, len(ratings))
	for _, tr := range ratings {
		resp = append(resp, typeRatingResponse{
			ID:              tr.ID,
			OrgID:           tr.OrgID,
			UserID:          tr.UserID,
			AircraftTypeID:  tr.AircraftTypeID,
			CertID:          tr.CertID,
			EndorsementDate: tr.EndorsementDate,
			Status:          tr.Status,
			CreatedAt:       tr.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func CreateUserTypeRating(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}
	var req typeRatingCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	aircraftTypeID, _ := uuid.Parse(req.AircraftTypeID)
	certID, _ := uuid.Parse(req.CertID)
	endorsementDate, _ := time.Parse(time.RFC3339, req.EndorsementDate)

	created, err := servicesReg.Certifications.CreateTypeRating(r.Context(), actor, services.TypeRatingCreateInput{
		UserID:          userID,
		AircraftTypeID:  aircraftTypeID,
		CertID:          certID,
		EndorsementDate: endorsementDate,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, typeRatingResponse{
		ID:              created.ID,
		OrgID:           created.OrgID,
		UserID:          created.UserID,
		AircraftTypeID:  created.AircraftTypeID,
		CertID:          created.CertID,
		EndorsementDate: created.EndorsementDate,
		Status:          created.Status,
		CreatedAt:       created.CreatedAt,
	})
}

func ListUserSkills(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}
	skills, err := servicesReg.Certifications.ListSkillsByUser(r.Context(), actor, userID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]skillResponse, 0, len(skills))
	for _, s := range skills {
		resp = append(resp, skillResponse{
			ID:                s.ID,
			OrgID:             s.OrgID,
			UserID:            s.UserID,
			SkillTypeID:       s.SkillTypeID,
			ProficiencyLevel:  s.ProficiencyLevel,
			QualificationDate: s.QualificationDate,
			ExpiryDate:        s.ExpiryDate,
			CreatedAt:         s.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func CreateUserSkill(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}
	var req skillCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	skillTypeID, _ := uuid.Parse(req.SkillTypeID)
	qualDate, _ := time.Parse(time.RFC3339, req.QualificationDate)
	var expiryDate *time.Time
	if req.ExpiryDate != "" {
		v, _ := time.Parse(time.RFC3339, req.ExpiryDate)
		expiryDate = &v
	}

	created, err := servicesReg.Certifications.CreateSkill(r.Context(), actor, services.SkillCreateInput{
		UserID:            userID,
		SkillTypeID:       skillTypeID,
		ProficiencyLevel:  req.ProficiencyLevel,
		QualificationDate: qualDate,
		ExpiryDate:        expiryDate,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, skillResponse{
		ID:                created.ID,
		OrgID:             created.OrgID,
		UserID:            created.UserID,
		SkillTypeID:       created.SkillTypeID,
		ProficiencyLevel:  created.ProficiencyLevel,
		QualificationDate: created.QualificationDate,
		ExpiryDate:        created.ExpiryDate,
		CreatedAt:         created.CreatedAt,
	})
}

func GetQualifiedMechanics(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}

	query := r.URL.Query()
	taskType := domain.TaskType(query.Get("task_type"))
	if taskType == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "task_type is required")
		return
	}

	var aircraftTypeID *uuid.UUID
	if atID := query.Get("aircraft_type_id"); atID != "" {
		parsed, err := uuid.Parse(atID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft_type_id")
			return
		}
		aircraftTypeID = &parsed
	}

	ids, err := servicesReg.Certifications.GetQualifiedMechanics(r.Context(), actor, taskType, aircraftTypeID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mechanic_ids": ids})
}

func CheckMechanicQualification(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid user id")
		return
	}

	query := r.URL.Query()
	taskType := domain.TaskType(query.Get("task_type"))
	if taskType == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "task_type is required")
		return
	}

	var aircraftTypeID *uuid.UUID
	if atID := query.Get("aircraft_type_id"); atID != "" {
		parsed, err := uuid.Parse(atID)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid aircraft_type_id")
			return
		}
		aircraftTypeID = &parsed
	}

	result, err := servicesReg.Certifications.CheckQualification(r.Context(), actor.OrgID, userID, taskType, aircraftTypeID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, qualificationCheckResponse{
		Qualified:           result.Qualified,
		MissingCerts:        result.MissingCerts,
		MissingTypeRatings:  result.MissingTypeRatings,
		MissingSkills:       result.MissingSkills,
		ExpiredCerts:        result.ExpiredCerts,
		InsufficientRecency: result.InsufficientRecency,
		Reasons:             result.Reasons,
	})
}

func CreateTaskSkillRequirement(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	servicesReg, ok := servicesFromRequest(r)
	if !ok || servicesReg.Certifications == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	var req requirementCreateRequest
	if err := decodeAndValidateJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}

	var aircraftTypeID *uuid.UUID
	if req.AircraftTypeID != nil {
		v, _ := uuid.Parse(*req.AircraftTypeID)
		aircraftTypeID = &v
	}
	var certTypeID *uuid.UUID
	if req.CertTypeID != nil {
		v, _ := uuid.Parse(*req.CertTypeID)
		certTypeID = &v
	}
	var skillTypeID *uuid.UUID
	if req.SkillTypeID != nil {
		v, _ := uuid.Parse(*req.SkillTypeID)
		skillTypeID = &v
	}

	created, err := servicesReg.Certifications.CreateRequirement(r.Context(), actor, services.RequirementCreateInput{
		TaskType:            domain.TaskType(req.TaskType),
		AircraftTypeID:      aircraftTypeID,
		CertTypeID:          certTypeID,
		SkillTypeID:         skillTypeID,
		MinProficiencyLevel: req.MinProficiencyLevel,
		IsCertifyingRole:    req.IsCertifyingRole,
		IsInspectionRole:    req.IsInspectionRole,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// --- Mapping Helpers ---

func mapCert(c domain.EmployeeCertification) certResponse {
	return certResponse{
		ID:                c.ID,
		OrgID:             c.OrgID,
		UserID:            c.UserID,
		CertTypeID:        c.CertTypeID,
		CertificateNumber: c.CertificateNumber,
		IssueDate:         c.IssueDate,
		ExpiryDate:        c.ExpiryDate,
		Status:            c.Status,
		VerifiedBy:        c.VerifiedBy,
		VerifiedAt:        c.VerifiedAt,
		DocumentURL:       c.DocumentURL,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}
