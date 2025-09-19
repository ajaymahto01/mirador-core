package models

import "time"

// TimeRange represents an absolute time window used by RCA requests.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Duration is a convenience helper (optional).
func (tr TimeRange) Duration() time.Duration {
	if tr.End.Before(tr.Start) {
		return 0
	}
	return tr.End.Sub(tr.Start)
}

// MetricsQL Models

// VictoriaMetricsResponse matches Prometheus-compatible JSON responses from VictoriaMetrics.
type VictoriaMetricsResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"` // can be instant vector, range vector, scalar, etc.
}

type MetricsQLQueryRequest struct {
    Query    string `json:"query" binding:"required"`
    Time     string `json:"time,omitempty"`
    Timeout  string `json:"timeout,omitempty"`
    TenantID string `json:"-"` // Set by middleware
    QueryLanguage     string  `json:"query_language,omitempty"` // e.g., "lucene", "promql"
    // Optional: include definitions and restrict label keys
    IncludeDefinitions *bool    `json:"include_definitions,omitempty"`
    DefinitionsMinimal *bool    `json:"definitions_minimal,omitempty"`
    LabelKeys          []string `json:"label_keys,omitempty"`
}

type MetricsQLRangeQueryRequest struct {
    Query    string `json:"query" binding:"required"`
    Start    string `json:"start" binding:"required"`
    End      string `json:"end" binding:"required"`
    Step     string `json:"step" binding:"required"`
    TenantID string `json:"-"`
    QueryLanguage     string  `json:"query_language,omitempty"`
    IncludeDefinitions *bool    `json:"include_definitions,omitempty"`
    DefinitionsMinimal *bool    `json:"definitions_minimal,omitempty"`
    LabelKeys          []string `json:"label_keys,omitempty"`
}

type MetricsQLQueryResult struct {
	Status        string      `json:"status"`
	Data          interface{} `json:"data"`
	SeriesCount   int         `json:"series_count"`
	ExecutionTime int64       `json:"execution_time_ms"`
}

type MetricsQLRangeQueryResult struct {
	Status         string      `json:"status"`
	Data           interface{} `json:"data"`
	DataPointCount int         `json:"data_point_count"`
}

type MetricsQLQueryResponse struct {
    Data          interface{} `json:"data"`
    ExecutionTime int64       `json:"execution_time"`
    Timestamp     time.Time   `json:"timestamp"`
    Definitions  interface{} `json:"definitions,omitempty"`
}

type SeriesRequest struct {
	Match    []string `json:"match[]"`
	Start    string   `json:"start,omitempty"`
	End      string   `json:"end,omitempty"`
	TenantID string   `json:"-"`
}

type LabelsRequest struct {
	Start    string   `json:"start,omitempty"`
	End      string   `json:"end,omitempty"`
	Match    []string `json:"match[],omitempty"`
	TenantID string   `json:"-"`
}

type LabelValuesRequest struct {
	Label    string   `json:"label"`
	Start    string   `json:"start,omitempty"`
	End      string   `json:"end,omitempty"`
	Match    []string `json:"match[],omitempty"`
	TenantID string   `json:"-"`
}

// LogsQL Models

type LogsQLQueryRequest struct {
    Query    string            `json:"query" form:"query"`
    Start    int64             `json:"start" form:"start"` // epoch (sec/ms/ns ok; service normalizes)
    End      int64             `json:"end" form:"end"`
    Limit    int               `json:"limit" form:"limit"`
    TenantID string            `json:"tenantId" form:"tenantId"`
    QueryLanguage string       `json:"query_language,omitempty"`
    Extra    map[string]string `json:"extra,omitempty" form:"-"` // passthrough flags (dedup, order, etc.)
}

type LogsQLQueryResult struct {
	Logs   []map[string]any `json:"logs,omitempty"`
	Fields []string         `json:"fields,omitempty"`
	Stats  map[string]any   `json:"stats,omitempty"`
}

type LogsQLResponse struct {
	Status string                   `json:"status"`
	Data   []map[string]interface{} `json:"data"`
}

type LogFieldsRequest struct {
	Query    string `json:"query,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	TenantID string `json:"-"`
}

type LogExportRequest struct {
    Query    string `json:"query" binding:"required"`
    Format   string `json:"format,omitempty"` // json, csv, parquet
    Start    int64  `json:"start,omitempty"` // epoch (sec/ms/ns ok; service normalizes)
    End      int64  `json:"end,omitempty"`
    Limit    int    `json:"limit,omitempty"`
    TenantID string `json:"-"`
    QueryLanguage string `json:"query_language,omitempty"`
}

type LogExportResult struct {
	ExportID      string    `json:"export_id"`
	Format        string    `json:"format"`
	RecordCount   int       `json:"record_count"`
	DownloadURL   string    `json:"download_url"`
	ExpiresAt     time.Time `json:"expires_at"`
	EstimatedSize string    `json:"estimated_size"`
}

// VictoriaTraces Models
type TraceSearchRequest struct {
    Service     string       `json:"service,omitempty"`
    Operation   string       `json:"operation,omitempty"`
    Tags        string       `json:"tags,omitempty"`
    MinDuration string       `json:"minDuration,omitempty"`
    MaxDuration string       `json:"maxDuration,omitempty"`
    Start       FlexibleTime `json:"start"`
    End         FlexibleTime `json:"end"`
    Limit       int          `json:"limit,omitempty"`
    TenantID    string       `json:"-"`
    Query       string       `json:"query,omitempty"`
    QueryLanguage string     `json:"query_language,omitempty"`
}

type TraceSearchResult struct {
	Traces     []map[string]interface{} `json:"traces"`
	Total      int                      `json:"total"`
	SearchTime int64                    `json:"search_time_ms"`
}

type Trace struct {
	TraceID   string                   `json:"traceID"`
	Spans     []map[string]interface{} `json:"spans"`
	Processes map[string]interface{}   `json:"processes"`
}

// AI Engine Models
type FractureAnalysisRequest struct {
	Component  string   `json:"component" binding:"required"`
	TimeRange  string   `json:"time_range" binding:"required"`
	ModelTypes []string `json:"model_types,omitempty"`
	TenantID   string   `json:"-"`
}

type FractureAnalysisResponse struct {
	Fractures        []*SystemFracture `json:"fractures"`
	ModelsUsed       []string          `json:"models_used"`
	ProcessingTimeMs int64             `json:"processing_time_ms"`
}

type RCAInvestigationRequest struct {
	IncidentID       string    `json:"incident_id" binding:"required"`
	Symptoms         []string  `json:"symptoms" binding:"required"`
	TimeRange        TimeRange `json:"time_range" binding:"required"`
	AffectedServices []string  `json:"affected_services,omitempty"`
	AnomalyThreshold float64   `json:"anomaly_threshold,omitempty"`
	TenantID         string    `json:"-"`
}

// Configuration Models
type DataSource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // metrics, logs, traces
	URL      string `json:"url"`
	Status   string `json:"status"` // connected, degraded, disconnected
	TenantID string `json:"tenant_id"`
}

type FractureListResponse struct {
	Fractures []*SystemFracture `json:"fractures"`
	Summary   FractureSummary   `json:"summary"`
}

type FractureSummary struct {
	Total         int           `json:"total"`
	HighRisk      int           `json:"high_risk"`
	MediumRisk    int           `json:"medium_risk"`
	LowRisk       int           `json:"low_risk"`
	AvgTimeToFail time.Duration `json:"avg_time_to_fail"`
}
