package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/cypher682/platform-cli/internal/health"
)

func colorStatus(s health.Status) string {
	switch s {
	case health.Healthy:
		return color.GreenString(string(s))
	case health.Degraded:
		return color.YellowString(string(s))
	case health.Unknown:
		return color.BlueString(string(s))
	case health.Skipped:
		return color.CyanString(string(s))
	default:
		return color.RedString(string(s))
	}
}

func PrintHealth(rs []health.Result) {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Service", "Port", "Status", "Latency", "URL"})
	for _, r := range rs {
		port := ""
		if r.Service.Port > 0 {
			port = fmt.Sprint(r.Service.Port)
		}
		url := r.Service.Health
		lat := ""
		if r.Latency > 0 {
			lat = r.Latency.Round(1e6).String()
		}
		t.Append([]string{r.Service.Name, port, colorStatus(r.Status), lat, url})
	}
	t.Render()
}
