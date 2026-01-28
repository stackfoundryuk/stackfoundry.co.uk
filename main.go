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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"stackfoundry.co.uk/components"
)

//go:embed public/*
var embeddedFiles embed.FS

var sesClient *ses.Client

const SenderEmail = "joe@stackfoundry.co.uk"

// AnalyticsMiddleware logs every request.
func AnalyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In tests, we might not want to log, but it's fine for now
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s | %v | %s", r.Method, r.URL.Path, time.Since(start), r.UserAgent())
	})
}

// setupRouter defines all routes and returns the handler.
// This allows us to test the routes without starting the main server.
func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// 1. Static Files
	publicFS, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		// In tests, if public folder isn't found/embedded, this might fail.
		// For a robust app, handle this gracefully or ensure test env has assets.
		log.Println("Warning: Public assets not found:", err)
	} else {
		mux.Handle("GET /css/", http.FileServer(http.FS(publicFS)))
		mux.Handle("GET /img/", http.FileServer(http.FS(publicFS)))
	}

	// 2. Exact Page Routes
	mux.Handle("GET /{$}", templ.Handler(components.Home()))
	mux.Handle("GET /privacy", templ.Handler(components.Privacy()))

	// 3. API Routes
	mux.HandleFunc("POST /api/contact", handleContact)

	// 4. Catch-All 404
	mux.Handle("/", templ.Handler(components.NotFound()))

	return mux
}

// Extracted handler for cleaner testing
func handleContact(w http.ResponseWriter, r *http.Request) {
	time.Sleep(100 * time.Millisecond)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	visitorEmail := r.FormValue("email")
	visitorSubject := r.FormValue("subject")
	visitorMessage := r.FormValue("message")

	// Only send if client is initialized (it won't be in tests, preventing errors)
	if sesClient != nil && visitorEmail != "" {
		err := sendEmail(r.Context(), visitorEmail, visitorSubject, visitorMessage)
		if err != nil {
			log.Printf("❌ SES ERROR: %v", err)
		} else {
			log.Printf("✅ Email sent successfully from %s", visitorEmail)
		}
	}

	components.ContactSuccess().Render(r.Context(), w)
}

func main() {
	// 0. Initialize AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Printf("⚠️ Warning: Unable to load AWS Config: %v", err)
	} else {
		sesClient = ses.NewFromConfig(cfg)
	}

	// 1. Setup Router
	mux := setupRouter()

	// 2. Middleware
	handler := AnalyticsMiddleware(mux)

	// 3. Start Server
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

func sendEmail(ctx context.Context, replyTo, subject, body string) error {
	input := &ses.SendEmailInput{
		Destination: &types.Destination{ToAddresses: []string{SenderEmail}},
		Message: &types.Message{
			Body: &types.Body{
				Text: &types.Content{Data: aws.String(fmt.Sprintf("From: %s\n\nMessage:\n%s", replyTo, body))},
				Html: &types.Content{Data: aws.String(fmt.Sprintf(`
					<h3>New Inquiry from StackFoundry</h3>
					<p><strong>From:</strong> %s</p>
					<p><strong>Subject:</strong> %s</p>
					<hr/>
					<p>%s</p>
				`, replyTo, subject, body))},
			},
			Subject: &types.Content{Data: aws.String("[StackFoundry] " + subject)},
		},
		Source:           aws.String(SenderEmail),
		ReplyToAddresses: []string{replyTo},
	}
	_, err := sesClient.SendEmail(ctx, input)
	return err
}
