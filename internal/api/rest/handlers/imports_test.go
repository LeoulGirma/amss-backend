package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

func newMultipartRequest(t *testing.T, method, path string, fields map[string]string, fileField, fileName string, fileBody []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	if fileField != "" {
		part, err := writer.CreateFormFile(fileField, fileName)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(fileBody); err != nil {
			t.Fatalf("write file body: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestCreateImport(t *testing.T) {
	orgID := uuid.New()
	importRepo := newFakeImportRepo()
	jobQueue := &fakeImportJobQueue{}
	importService := &services.ImportService{Imports: importRepo, Jobs: jobQueue}
	handler := ImportHandler{Service: importService, StorageDir: t.TempDir()}

	req := newMultipartRequest(t, http.MethodPost, "/api/v1/imports", map[string]string{
		"type": string(domain.ImportTypeAircraft),
	}, "file", "aircraft.csv", []byte("tail_number,model,capacity_slots\nN123,737,10\n"))
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler.CreateImport(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rr.Code)
	}
	var resp importResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Type != domain.ImportTypeAircraft {
		t.Fatalf("expected import type aircraft, got %s", resp.Type)
	}
	if resp.FileName != "aircraft.csv" {
		t.Fatalf("expected filename aircraft.csv, got %s", resp.FileName)
	}
	jobQueue.mu.Lock()
	queued := len(jobQueue.queued)
	jobQueue.mu.Unlock()
	if queued != 1 {
		t.Fatalf("expected 1 queued job, got %d", queued)
	}
}

func TestCreateImportUnauthorized(t *testing.T) {
	req := newMultipartRequest(t, http.MethodPost, "/api/v1/imports", map[string]string{
		"type": string(domain.ImportTypeAircraft),
	}, "file", "aircraft.csv", []byte("tail_number,model,capacity_slots\nN123,737,10\n"))

	rr := httptest.NewRecorder()
	ImportHandler{Service: &services.ImportService{}}.CreateImport(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestCreateImportMissingFile(t *testing.T) {
	orgID := uuid.New()
	importRepo := newFakeImportRepo()
	importService := &services.ImportService{Imports: importRepo}
	handler := ImportHandler{Service: importService, StorageDir: t.TempDir()}

	req := newMultipartRequest(t, http.MethodPost, "/api/v1/imports", map[string]string{
		"type": string(domain.ImportTypeAircraft),
	}, "", "", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)

	rr := httptest.NewRecorder()
	handler.CreateImport(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetImport(t *testing.T) {
	orgID := uuid.New()
	importRepo := newFakeImportRepo()
	importService := &services.ImportService{Imports: importRepo}
	handler := ImportHandler{Service: importService}

	imp := domain.Import{
		ID:        uuid.New(),
		OrgID:     orgID,
		Type:      domain.ImportTypeAircraft,
		Status:    domain.ImportStatusPending,
		FileName:  "aircraft.csv",
		FilePath:  "tmp",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = importRepo.Create(context.Background(), imp)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/imports/"+imp.ID.String(), nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", imp.ID.String())

	rr := httptest.NewRecorder()
	handler.GetImport(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp importResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != imp.ID {
		t.Fatalf("expected import id %s, got %s", imp.ID, resp.ID)
	}
}

func TestListImportRows(t *testing.T) {
	orgID := uuid.New()
	importRepo := newFakeImportRepo()
	rowRepo := newFakeImportRowRepo()
	importService := &services.ImportService{Imports: importRepo, Rows: rowRepo}
	handler := ImportHandler{Service: importService}

	impID := uuid.New()
	row := domain.ImportRow{
		ID:        uuid.New(),
		OrgID:     orgID,
		ImportID:  impID,
		RowNumber: 2,
		Status:    domain.ImportRowInvalid,
		Errors:    []string{"missing tail_number"},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_ = rowRepo.Create(context.Background(), row)

	req := newJSONRequest(t, http.MethodGet, "/api/v1/imports/"+impID.String()+"/rows", nil)
	req = withPrincipal(req, orgID, domain.RoleScheduler)
	req = withRouteParam(req, "id", impID.String())

	rr := httptest.NewRecorder()
	handler.ListImportRows(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []importRowResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 import row, got %d", len(resp))
	}
	if resp[0].ID != row.ID {
		t.Fatalf("expected row id %s, got %s", row.ID, resp[0].ID)
	}
}
