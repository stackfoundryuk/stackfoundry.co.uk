package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
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

type ContextKey string

const SessionKey ContextKey = "session_id"

// LoggerMiddleware: Tracks sessions, filters bots, and adds Security Headers
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// --- SECURITY HEADERS ---
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// 1. SESSION ID
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			b := make([]byte, 3)
			if _, err := rand.Read(b); err == nil {
				sessionID = hex.EncodeToString(b)
			} else {
				sessionID = "unknown"
			}
		}

		// 2. DETECT BOTS
		ua := r.UserAgent()
		isBot := strings.Contains(ua, "bot") ||
			strings.Contains(ua, "validation") ||
			strings.Contains(ua, "spider")

		// 3. CAPTURE HTMX CONTEXT
		clickedElement := r.Header.Get("HX-Trigger")
		if clickedElement == "" {
			clickedElement = "page_load"
		}
		currentURL := r.Header.Get("HX-Current-URL")

		// 4. PREPARE CONTEXT & LOGGER
		ctx := context.WithValue(r.Context(), SessionKey, sessionID)
		r = r.WithContext(ctx)

		logger := slog.Default().With(
			slog.String("session", sessionID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("click", clickedElement),
		)

		next.ServeHTTP(w, r)

		// 5. LOG (Filter noise)
		duration := time.Since(start)
		if isBot {
			logger.Info("bot_traffic",
				slog.String("ua", ua),
				slog.Duration("dur", duration),
			)
		} else {
			logger.Info("human_traffic",
				slog.String("url_context", currentURL),
				slog.Duration("dur", duration),
				slog.String("ua", ua),
			)
		}
	})
}

// CacheControlMiddleware: Forces browsers to cache static assets for 1 year
func CacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// "public" = CDN can cache it. "immutable" = content never changes.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

// RenderHTML: Helper for correct headers
func RenderHTML(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	// Inject Session ID on full page loads so HTMX picks it up
	if r.Header.Get("HX-Request") == "" {
		sessionID, _ := r.Context().Value(SessionKey).(string)
		w.Header().Set("X-Session-ID", sessionID)
	}

	component.Render(r.Context(), w)
}

// serveEmbeddedFile is a helper to serve a single file from the embedded FS
// with the correct Content-Type.
func serveEmbeddedFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, path string, contentType string) {
	file, err := fsys.Open(path)
	if err != nil {
		slog.Error("file_missing", slog.String("path", path), slog.Any("error", err))
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	io.Copy(w, file)
}

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	publicFS, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		slog.Error("assets_missing", slog.Any("error", err))
	} else {
		// 1. STATIC ASSETS (Images & CSS) -> CACHED 1 YEAR
		assetHandler := http.FileServer(http.FS(publicFS))
		mux.Handle("GET /css/", CacheControlMiddleware(assetHandler))
		mux.Handle("GET /img/", CacheControlMiddleware(assetHandler))

		// 2. JS ASSETS (HTMX) -> CACHED 1 YEAR
		mux.Handle("GET /js/", CacheControlMiddleware(assetHandler))

		// 3. SITEMAP
		mux.HandleFunc("GET /sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
			serveEmbeddedFile(w, r, publicFS, "sitemap.xml", "application/xml")
		})

		// 4. ROBOTS.TXT
		mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
			serveEmbeddedFile(w, r, publicFS, "robots.txt", "text/plain")
		})

		// 5. LLMS.TXT (The main AI directive file)
		mux.HandleFunc("GET /llms.txt", func(w http.ResponseWriter, r *http.Request) {
			serveEmbeddedFile(w, r, publicFS, "llms.txt", "text/plain; charset=utf-8")
		})

		// 6. LLM.TXT (Redirect to plural standard)
		mux.HandleFunc("GET /llm.txt", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/llms.txt", http.StatusMovedPermanently)
		})
	}

	// Page Routes
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		sessionID, _ := r.Context().Value(SessionKey).(string)
		RenderHTML(w, r, components.Home(sessionID))
	})

	mux.HandleFunc("GET /privacy", func(w http.ResponseWriter, r *http.Request) {
		sessionID, _ := r.Context().Value(SessionKey).(string)
		RenderHTML(w, r, components.Privacy(sessionID))
	})

	// API Routes
	mux.HandleFunc("POST /api/contact", handleContact)

	// 404 Handler
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		sessionID, _ := r.Context().Value(SessionKey).(string)
		RenderHTML(w, r, components.NotFound(sessionID))
	}))

	return mux
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	visitorEmail := r.FormValue("email")
	visitorSubject := r.FormValue("subject")
	visitorMessage := r.FormValue("message")

	slog.Info("contact_attempt", slog.String("email", visitorEmail))

	if sesClient != nil && visitorEmail != "" {
		err := sendEmail(r.Context(), visitorEmail, visitorSubject, visitorMessage)
		if err != nil {
			slog.Error("ses_failure", slog.Any("error", err))
		} else {
			slog.Info("ses_success", slog.String("recipient", visitorEmail))
		}
	}

	RenderHTML(w, r, components.ContactSuccess())
}

func main() {
	// Configure JSON Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		slog.Warn("aws_config_failed", slog.Any("error", err))
	} else {
		sesClient = ses.NewFromConfig(cfg)
	}

	mux := setupRouter()
	handler := LoggerMiddleware(mux)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		slog.Info("server_starting", slog.String("mode", "lambda_v1"))
		adapter := httpadapter.New(handler)
		lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.ProxyWithContext(ctx, req)
		})
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		slog.Info("server_starting", slog.String("mode", "local"), slog.String("port", port))
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
