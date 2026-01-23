package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
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

// AnalyticsMiddleware logs every request to stdout/CloudWatch.
// This is GDPR compliant as it does not store cookies or personal identifiers on the client.
func AnalyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Pass the request down the chain
		next.ServeHTTP(w, r)

		// Log the hit (CloudWatch will capture this)
		// Format: [METHOD] Path | Duration | UserAgent
		log.Printf("[%s] %s | %v | %s", r.Method, r.URL.Path, time.Since(start), r.UserAgent())
	})
}

func main() {
	mux := http.NewServeMux()

	// 1. Static Files (CSS/Images)
	// We use "GET /css/" to ensure the catch-all doesn't swallow these
	publicFS, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /css/", http.FileServer(http.FS(publicFS)))
	mux.Handle("GET /img/", http.FileServer(http.FS(publicFS)))

	// 2. Exact Page Routes
	// "GET /{$}" matches ONLY the root domain (Go 1.22+)
	mux.Handle("GET /{$}", templ.Handler(components.Home()))

	// Privacy Policy
	mux.Handle("GET /privacy", templ.Handler(components.Privacy()))

	// 3. API Routes
	// Contact Form Logic
	mux.HandleFunc("GET /api/contact", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond) // UX Delay
		sse := datastar.NewSSE(w, r)
		if err := sse.MergeFragmentTempl(components.ContactSuccess()); err != nil {
			sse.ConsoleError(err)
		}
	})

	// Hero Animation Stream (Calls the function in animation.go)
	mux.HandleFunc("GET /api/hero-stream", MatrixRainHandler)

	// 4. Catch-All 404
	// This matches any path not defined above (e.g. /invalid-page)
	mux.Handle("/", templ.Handler(components.NotFound()))

	// 5. Wrap everything in Analytics
	// This ensures 404s and valid pages are all logged
	handler := AnalyticsMiddleware(mux)

	// 6. Start Server (AWS or Local)
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("☁️  StackFoundry: Running in AWS Lambda Mode")
		adapter := httpadapter.New(handler)
		lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.ProxyWithContext(ctx, req)
		})
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		fmt.Printf("⚡ StackFoundry running on http://localhost:%s\n", port)
		http.ListenAndServe(":"+port, handler)
	}
}
