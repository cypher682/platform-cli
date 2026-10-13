package cmd

import (
	"github.com/cypher682/platform-cli/internal/health"
	outpkg "github.com/cypher682/platform-cli/internal/output"
)

func printResults(rs any) {
	switch v := rs.(type) {
	case []health.Result:
		if globalOpts.Format == outpkg.JSON {
			_ = outFmt.WriteJSON(v)
			return
		}
		outpkg.PrintHealth(v)
	default:
		if globalOpts.Format == outpkg.JSON {
			_ = outFmt.WriteJSON(v)
			return
		}
		_ = outFmt.Print(v)
	}
}
