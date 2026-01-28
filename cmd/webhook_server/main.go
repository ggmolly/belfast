package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const repoPath = "/home/molly/Documents/belfast"

var (
	defaultPNGPath = "/media/brooklyn/cdn/belfast/implem.png"
	defaultFont    = "monospace"
)

type jobRunner struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	cmd    *exec.Cmd
}

func (runner *jobRunner) start() {
	runner.mu.Lock()
	runner.cancelCurrentLocked()
	ctx, cancel := context.WithCancel(context.Background())
	runner.cancel = cancel
	go runner.run(ctx)
	runner.mu.Unlock()
}

func (runner *jobRunner) cancelCurrentLocked() {
	if runner.cancel != nil {
		runner.cancel()
	}
	if runner.cmd != nil {
		terminateProcessGroup(runner.cmd)
	}
}

func (runner *jobRunner) run(ctx context.Context) {
	defer func() {
		runner.mu.Lock()
		runner.cmd = nil
		runner.mu.Unlock()
	}()

	pngPath := getenvDefault("WEBHOOK_PNG_PATH", defaultPNGPath)
	fontFamily := getenvDefault("WEBHOOK_FONT_FAMILY", defaultFont)
	if err := os.MkdirAll(filepath.Dir(pngPath), 0o755); err != nil {
		return
	}
	if err := runner.runCommand(ctx, []string{"git", "pull", "--ff-only"}, "git pull"); err != nil {
		return
	}
	goCmd := []string{
		"go",
		"run",
		"./cmd/packet_progress",
		"png",
		"-png-scale",
		"1.5",
		"-font-family",
		fontFamily,
		"-out-png",
		pngPath,
	}
	_ = runner.runCommand(ctx, goCmd, "generate png")
}

func (runner *jobRunner) runCommand(ctx context.Context, cmdArgs []string, label string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = repoPath
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	env := os.Environ()
	if os.Getenv("WEBHOOK_SSH_KEY") != "" && cmdArgs[0] == "git" {
		sshCmd := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", os.Getenv("WEBHOOK_SSH_KEY"))
		env = append(env, "GIT_SSH_COMMAND="+sshCmd)
	}
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	runner.mu.Lock()
	if ctx.Err() != nil {
		runner.mu.Unlock()
		return ctx.Err()
	}
	runner.cmd = cmd
	runner.mu.Unlock()

	if err := cmd.Start(); err != nil {
		return err
	}

	output := bytes.Buffer{}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if ctx.Err() != nil {
			terminateProcessGroup(cmd)
			return ctx.Err()
		}
		output.Write(scanner.Bytes())
		output.WriteByte('\n')
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			return fmt.Errorf("%s failed: %w\n%s", label, err, output.String())
		}
	case <-ctx.Done():
		terminateProcessGroup(cmd)
		<-done
		return ctx.Err()
	}

	return nil
}

func terminateProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}

type webhookHandler struct {
	runner *jobRunner
}

func (handler *webhookHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/health":
		handler.sendJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
	case request.Method == http.MethodPost && request.URL.Path == "/webhook":
		handler.handleWebhook(writer, request)
	default:
		http.NotFound(writer, request)
	}
}

func (handler *webhookHandler) handleWebhook(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "failed to read payload", http.StatusBadRequest)
		return
	}
	if !verifySignature(request.Header.Get("X-Hub-Signature-256"), payload, writer) {
		return
	}
	handler.runner.start()
	handler.sendJSON(writer, http.StatusOK, map[string]string{"status": "queued"})
}

func (handler *webhookHandler) sendJSON(writer http.ResponseWriter, status int, payload map[string]string) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Content-Length", strconv.Itoa(len(body)))
	writer.WriteHeader(status)
	_, _ = writer.Write(body)
}

func verifySignature(signature string, payload []byte, writer http.ResponseWriter) bool {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		http.Error(writer, "WEBHOOK_SECRET is not set", http.StatusInternalServerError)
		return false
	}
	if !strings.HasPrefix(signature, "sha256=") {
		http.Error(writer, "missing signature", http.StatusUnauthorized)
		return false
	}
	provided := strings.TrimPrefix(signature, "sha256=")
	expectedMAC := hmac.New(sha256.New, []byte(secret))
	_, _ = expectedMAC.Write(payload)
	expected := expectedMAC.Sum(nil)
	providedBytes, err := hex.DecodeString(provided)
	if err != nil {
		http.Error(writer, "invalid signature", http.StatusUnauthorized)
		return false
	}
	if !hmac.Equal(expected, providedBytes) {
		http.Error(writer, "invalid signature", http.StatusUnauthorized)
		return false
	}
	return true
}

func getenvDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	host := getenvDefault("WEBHOOK_HOST", "0.0.0.0")
	portRaw := getenvDefault("WEBHOOK_PORT", "8080")
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		panic(err)
	}

	runner := &jobRunner{}
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: &webhookHandler{runner: runner},
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
