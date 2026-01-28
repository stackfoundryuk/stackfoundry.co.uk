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

	"stackfoundry.co.uk/components"
)

//go:embed public/*
var embeddedFiles embed.FS

// AnalyticsMiddleware logs every request.
func AnalyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s | %v | %s", r.Method, r.URL.Path, time.Since(start), r.UserAgent())
	})
}

func main() {
	mux := http.NewServeMux()

	// 1. Static Files (CSS/Images)
	publicFS, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /css/", http.FileServer(http.FS(publicFS)))
	mux.Handle("GET /img/", http.FileServer(http.FS(publicFS)))

	// 2. Exact Page Routes
	mux.Handle("GET /{$}", templ.Handler(components.Home()))
	mux.Handle("GET /privacy", templ.Handler(components.Privacy()))

	// 3. API Routes
	// Contact Form Logic (HTMX POST)
	mux.HandleFunc("POST /api/contact", func(w http.ResponseWriter, r *http.Request) {
		// Artificial delay to show off the "Initialize Sequence" animation
		time.Sleep(800 * time.Millisecond)

		// Here is where you would normally parse the form:
		// r.ParseForm()
		// email := r.FormValue("email")
		// ... Send to SES ...

		// Render the Success Component to replace the form
		components.ContactSuccess().Render(r.Context(), w)
	})

	// 4. Catch-All 404
	mux.Handle("/", templ.Handler(components.NotFound()))

	// 5. Wrap everything in Analytics
	handler := AnalyticsMiddleware(mux)

	// 6. Start Server (AWS or Local)
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("stackfoundry.co.uk: Running in AWS Lambda Mode")
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
