package monitor

type CheckResult struct {
	URL        string  `json:"url"`
	StatusCode int     `json:"status_code"`
	DurationMS float64 `json:"duration_ms"`
	Healthy    bool    `json:"healthy"`
	Error      *string `json:"error"`
}

type Summary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Errors    int `json:"errors"`
}

type Report struct {
	Results []CheckResult `json:"results"`
	Summary Summary       `json:"summary"`
}
