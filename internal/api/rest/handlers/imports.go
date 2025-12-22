package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ImportHandler struct {
	Service    *services.ImportService
	StorageDir string
}

type importResponse struct {
	ID        uuid.UUID           `json:"id"`
	OrgID     uuid.UUID           `json:"org_id"`
	Type      domain.ImportType   `json:"type"`
	Status    domain.ImportStatus `json:"status"`
	FileName  string              `json:"file_name"`
	CreatedBy uuid.UUID           `json:"created_by"`
	Summary   map[string]any      `json:"summary,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type importRowResponse struct {
	ID        uuid.UUID              `json:"id"`
	RowNumber int                    `json:"row_number"`
	Status    domain.ImportRowStatus `json:"status"`
	Errors    []string               `json:"errors,omitempty"`
	Raw       map[string]any         `json:"raw,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func (h ImportHandler) CreateImport(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	if h.Service == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid multipart form")
		return
	}
	importType := r.FormValue("type")
	if importType == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "type is required")
		return
	}
	orgID := actor.OrgID
	if actor.IsAdmin() {
		if orgVal := r.FormValue("org_id"); orgVal != "" {
			parsed, err := uuid.Parse(orgVal)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			orgID = parsed
		}
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "file is required")
		return
	}
	defer file.Close()

	dir := h.StorageDir
	if dir == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "failed to prepare storage")
		return
	}
	impID := uuid.New()
	filePath := filepath.Join(dir, impID.String()+".csv")
	dest, err := os.Create(filePath)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "failed to create file")
		return
	}
	if _, err := io.Copy(dest, file); err != nil {
		_ = dest.Close()
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "failed to store file")
		return
	}
	if err := dest.Close(); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "failed to store file")
		return
	}

	created, err := h.Service.Create(r.Context(), actor, services.ImportCreateInput{
		ID:        impID,
		OrgID:     &orgID,
		Type:      domain.ImportType(importType),
		FileName:  header.Filename,
		FilePath:  filePath,
		CreatedBy: actor.UserID,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, mapImport(created))
}

func (h ImportHandler) GetImport(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	if h.Service == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid import id")
		return
	}
	orgID := actor.OrgID
	if actor.IsAdmin() {
		if org := r.URL.Query().Get("org_id"); org != "" {
			parsed, err := uuid.Parse(org)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			orgID = parsed
		}
	}
	imp, err := h.Service.Get(r.Context(), actor, orgID, id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapImport(imp))
}

func (h ImportHandler) ListImportRows(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
		return
	}
	if h.Service == nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL", "service unavailable")
		return
	}
	importID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid import id")
		return
	}
	filter := ports.ImportRowFilter{
		OrgID:    actor.OrgID,
		ImportID: importID,
	}
	if actor.IsAdmin() {
		if org := r.URL.Query().Get("org_id"); org != "" {
			orgID, err := uuid.Parse(org)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid org_id")
				return
			}
			filter.OrgID = orgID
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		value := domain.ImportRowStatus(status)
		filter.Status = &value
	} else {
		value := domain.ImportRowInvalid
		filter.Status = &value
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		value, err := parseInt(limit)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid limit")
			return
		}
		filter.Limit = value
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		value, err := parseInt(offset)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid offset")
			return
		}
		filter.Offset = value
	}

	rows, err := h.Service.ListRows(r.Context(), actor, filter)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	resp := make([]importRowResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, mapImportRow(row))
	}
	writeJSON(w, http.StatusOK, resp)
}

func mapImport(imp domain.Import) importResponse {
	return importResponse{
		ID:        imp.ID,
		OrgID:     imp.OrgID,
		Type:      imp.Type,
		Status:    imp.Status,
		FileName:  imp.FileName,
		CreatedBy: imp.CreatedBy,
		Summary:   imp.Summary,
		CreatedAt: imp.CreatedAt,
		UpdatedAt: imp.UpdatedAt,
	}
}

func mapImportRow(row domain.ImportRow) importRowResponse {
	return importRowResponse{
		ID:        row.ID,
		RowNumber: row.RowNumber,
		Status:    row.Status,
		Errors:    row.Errors,
		Raw:       row.Raw,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
