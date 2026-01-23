package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	datastar "github.com/starfederation/datastar/sdk/go"
	"stackfoundry.co.uk/components"
)

func MatrixRainHandler(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(60 * time.Second)

	const TotalCells = 384

	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout:
			return
		case <-ticker.C:
			// Pick ANY cell on the screen
			cellID := rand.Intn(TotalCells)

			val := fmt.Sprintf("%02X", rand.Intn(255))

			if err := sse.MergeFragmentTempl(components.HexCell(cellID, val, true)); err != nil {
				return
			}
		}
	}
}
