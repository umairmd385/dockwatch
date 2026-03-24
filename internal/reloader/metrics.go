package reloader

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ReloadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reloader_actions_total",
		Help: "The total number of container reload actions triggered",
	}, []string{"container_name", "action", "status"})

	FileChangesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reloader_file_changes_total",
		Help: "The total number of file changes detected",
	}, []string{"path"})
)
