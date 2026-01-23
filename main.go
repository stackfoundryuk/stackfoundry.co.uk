package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	datastar "github.com/starfederation/datastar/sdk/go"

	"stackfoundry.co.uk/components"
)

//go:embed public/*
var embeddedFiles embed.FS

func main() {
	mux := http.NewServeMux()

	// 1. Static Files (CSS)
	publicFS, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/css/", http.FileServer(http.FS(publicFS)))

	// 2. Home Page
	mux.Handle("/", templ.Handler(components.Home()))

	// 3. Contact API
	mux.HandleFunc("/api/contact", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond) // Simulate network
		sse := datastar.NewSSE(w, r)
		if err := sse.MergeFragmentTempl(components.ContactSuccess()); err != nil {
			sse.ConsoleError(err)
		}
	})

	// 4. Hero Animation Stream (The Matrix Rain)
	mux.HandleFunc("/api/hero-stream", func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)

		// Update speed: 50ms = rapid tech feel
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		// Safety: Stop animating after 30s to save CPU/Battery
		timeout := time.After(30 * time.Second)

		for {
			select {
			case <-r.Context().Done():
				return // Client closed tab
			case <-timeout:
				return // Time limit reached
			case <-ticker.C:
				// Pick random cell (0-95)
				cellID := rand.Intn(96)

				// Generate random hex (e.g. "AF")
				val := fmt.Sprintf("%02X", rand.Intn(255))

				// Send the update to the client
				// highlight=true makes it flash briefly
				if err := sse.MergeFragmentTempl(components.HexCell(cellID, val, true)); err != nil {
					return
				}
			}
		}
	})

	// 5. Start Server (Lambda or Local)
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("☁️  Running in AWS Lambda Mode")
		adapter := httpadapter.New(mux)

		lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.ProxyWithContext(ctx, req)
		})
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		fmt.Printf("⚡ StackFoundry running on http://localhost:%s\n", port)
		http.ListenAndServe(":"+port, mux)
	}
}
