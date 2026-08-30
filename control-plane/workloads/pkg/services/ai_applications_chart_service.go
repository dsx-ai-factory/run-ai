package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
)

const chartFormatErrorMessage = "Only .tgz Helm chart packages are supported."

// chartDownloadTempFileTTL bounds how long the temp file backing a download response lives. The
// generated octet-stream handler reads the file into memory at the start of the response, so any
// value comfortably above that read is safe; we keep it generous for slow clients.
const chartDownloadTempFileTTL = 5 * time.Minute

// UploadApplicationChart validates and stores an uploaded Helm chart archive, returning a chartId
// that can be referenced as spec.chartRef.chartId when creating an AI application. The generated
// controller hands us an already-closed temp file, so we read and remove it by path.
func (s *AIApplicationsService) UploadApplicationChart(ctx context.Context, chart *os.File) (api.ImplResponse, error) {
	if chart == nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "chart file is required.", nil)), nil
	}
	chartPath := chart.Name()
	filename := originalChartFilename(chartPath)
	defer func() { _ = os.Remove(chartPath) }()

	// No ApplicationSubmission gate here: the other chart sources (custom URL, NGC catalog) have no
	// pre-create step that checks it, and an uploaded chart is inert until CreateApplication references
	// its chartId, where the existing gate applies. Keeping the check only there avoids a second,
	// source-specific gate the other sources don't have.
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	conf := runai_config.Get[*config.ChartUploadConfig]()

	info, statErr := os.Stat(chartPath)
	if statErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read uploaded chart %q", filename), nil).WithDetails(statErr)), nil
	}
	if info.Size() == 0 {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "Upload a file", nil)), nil
	}
	if info.Size() > conf.MaxUploadBytes {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("File exceeds the %d MB limit.", conf.MaxUploadBytes/(1024*1024)), nil)), nil
	}

	data, readErr := os.ReadFile(chartPath)
	if readErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read uploaded chart %q", filename), nil).WithDetails(readErr)), nil
	}
	if !isGzip(data) {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, chartFormatErrorMessage, nil)), nil
	}

	upload := &models.ApplicationChartUpload{
		ID:         uuid.New(),
		TenantId:   fmt.Sprint(tenantId),
		ChartBytes: data,
		CreatedAt:  time.Now(),
	}
	if dbErr := dao.CreateApplicationChartUpload(ctx, upload, filename); dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	return api.Response(http.StatusOK, api.AiApplicationChartUploadResponse{ChartId: upload.ID.String()}), nil
}

// DownloadApplicationChart streams the stored chart archive to the cluster. The endpoint is
// JWT-authenticated (the application controller sends the per-cluster bearer token) and scoped to the
// caller's tenant, so a chart is only reachable by its owning tenant. Transient retention still bounds
// how long a chart lives.
func (s *AIApplicationsService) DownloadApplicationChart(ctx context.Context, chartId string) (api.ImplResponse, error) {
	id, err := uuid.Parse(chartId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "chartId must be a valid uuid.", err)), nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	upload, apiErr := dao.GetApplicationChartUploadByID(ctx, id, fmt.Sprint(tenantId))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tempPath, writeErr := writeTempChart(upload.ChartBytes)
	if writeErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to prepare chart for download", nil).WithDetails(writeErr)), nil
	}
	// The generated controller streams this file after the method returns, so removal cannot be
	// deferred here; schedule it once the synchronous response has certainly completed.
	time.AfterFunc(chartDownloadTempFileTTL, func() { _ = os.Remove(tempPath) })

	return api.Response(http.StatusOK, tempPath), nil
}

func writeTempChart(data []byte) (string, error) {
	file, err := os.CreateTemp("", "runai-chart-*.tgz")
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(data); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

// originalChartFilename recovers the uploaded file's name from the generated router's temp path. That
// path is "<original name>.<random>" (os.CreateTemp with a "*" suffix), so trimming at the final dot
// yields the name the user picked, used to make chart errors recognizable.
func originalChartFilename(tempPath string) string {
	base := filepath.Base(tempPath)
	if idx := strings.LastIndex(base, "."); idx > 0 {
		return base[:idx]
	}
	return base
}

// isGzip reports whether data is a valid gzip stream (a Helm .tgz is gzip-compressed tar). It reads
// the whole stream so the gzip CRC/length trailer is validated, not just the header; the caller has
// already size-capped the input, so draining to io.Discard carries no decompression-bomb risk.
func isGzip(data []byte) bool {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return false
	}
	if _, err := io.Copy(io.Discard, reader); err != nil {
		_ = reader.Close()
		return false
	}
	return reader.Close() == nil
}
