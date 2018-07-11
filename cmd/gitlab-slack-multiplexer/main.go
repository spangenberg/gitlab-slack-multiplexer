package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/spangenberg/gitlab-slack-multiplexer/src/version"
)

type key int

type responseMessage struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

const (
	requestIDKey key = 0
)

var (
	listenAddr string
	gitlabURL  string

	healthy int32

	re = regexp.MustCompile(`\A\s*([-\w]+(?:\/[-\w]+){1,})\s*(.*)\z`)

	channelNotLinkedReponse = ephemeralResponse("Please specify a namespaced project as first parameter! This channel isn't linked to a project.")
	directMessageResponse   = ephemeralResponse("Please specify a namespaced project as first parameter!")
	privateGroupResponse    = ephemeralResponse("Please specify a namespaced project as first parameter! Private channels don't yet support project binding.")
	projectNotFoundResponse = ephemeralResponse("Ops! Looks like you're trying to access a project which hasn't been setup yet with a slack integration.")
	unkownResponse          = ephemeralResponse("Ops! Looks like something went wrong talking to GitLab!")
)

func ephemeralResponse(text string) string {
	response := responseMessage{
		ResponseType: "ephemeral",
		Text:         text,
	}

	b, _ := json.Marshal(response)

	return string(b)
}

func getNamespacedProjectByChannel(channelName string) string {
	// TODO: Implement channel lookup logic
	return ""
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, "show version")
	flag.StringVar(&gitlabURL, "gitlab-url", os.Getenv("GITLAB_URL"), "GitLab URL")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "server listen address")
	flag.Parse()

	if versionFlag != false {
		version.PrintAndExit()
	}

	if gitlabURL == "" {
		log.Fatal("GitLab URL configuration missing.")
	}

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := http.NewServeMux()
	router.Handle("/healthz", healthz())
	router.Handle("/slack/command", slackCommand())

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

func proxy(namespacedProject string, form url.Values) (string, error) {
	requestURL := fmt.Sprintf("%s/api/v4/projects/%s/services/slack_slash_commands/trigger", gitlabURL, url.QueryEscape(namespacedProject))

	body := bytes.NewBufferString(form.Encode())
	rsp, err := http.Post(requestURL, "application/x-www-form-urlencoded", body)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	switch rsp.StatusCode {
	case 404:
		return projectNotFoundResponse, nil
	case 200:
		body_byte, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return "", err
		}
		return string(body_byte), nil
	default:
		return unkownResponse, nil
	}
}

func slackCommand() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		r.ParseForm()

		var namespacedProject string
		result := re.FindStringSubmatch(r.PostForm.Get("text"))
		if len(result) > 0 {
			namespacedProject = result[1]
			r.PostForm.Set("text", result[2])
		} else {
			switch channelName := r.PostForm.Get("channel_name"); channelName {
			case "directmessage":
				io.WriteString(w, directMessageResponse)
				return
			case "privategroup":
				io.WriteString(w, privateGroupResponse)

				return
			default:
				if namespacedProject = getNamespacedProjectByChannel(channelName); namespacedProject == "" {
					io.WriteString(w, channelNotLinkedReponse)

					return
				}
			}
		}

		resp, err := proxy(namespacedProject, r.PostForm)
		if err != nil {
			log.Println(err)
			io.WriteString(w, unkownResponse)

			return
		}

		io.WriteString(w, resp)
	})
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
