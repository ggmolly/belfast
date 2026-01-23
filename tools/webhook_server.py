#!/usr/bin/env python3
import hashlib
import hmac
import json
import os
import signal
import subprocess
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


REPO_PATH = "/home/molly/Documents/belfast"
PNG_PATH = os.environ.get("WEBHOOK_PNG_PATH", "/media/brooklyn/cdn/belfast/implem.png")
FONT_FAMILY = os.environ.get("WEBHOOK_FONT_FAMILY", "monospace")
GO_CMD = [
    "go",
    "run",
    "./cmd/packet_progress",
    "png",
    "-png-scale",
    "1.5",
    "-font-family",
    FONT_FAMILY,
    "-out-png",
    PNG_PATH,
]


class JobRunner:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._thread: threading.Thread | None = None
        self._cancel = threading.Event()
        self._process: subprocess.Popen | None = None

    def start(self) -> None:
        with self._lock:
            self._cancel_current_locked()
            self._cancel = threading.Event()
            thread = threading.Thread(target=self._run, daemon=True)
            self._thread = thread
            thread.start()

    def _cancel_current_locked(self) -> None:
        if self._thread and self._thread.is_alive():
            self._cancel.set()
            self._terminate_process_locked()

    def _terminate_process_locked(self) -> None:
        if self._process is None:
            return
        try:
            os.killpg(self._process.pid, signal.SIGTERM)
        except ProcessLookupError:
            pass
        self._process = None

    def _run(self) -> None:
        try:
            os.makedirs(os.path.dirname(PNG_PATH), exist_ok=True)
            self._run_command(["git", "pull", "--ff-only"], "git pull")
            self._run_command(GO_CMD, "generate png")
        finally:
            with self._lock:
                self._process = None

    def _run_command(self, cmd: list[str], label: str) -> None:
        if self._cancel.is_set():
            return
        with self._lock:
            if self._cancel.is_set():
                return
            self._process = subprocess.Popen(
                cmd,
                cwd=REPO_PATH,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
                preexec_fn=os.setsid,
            )
        output = []
        while True:
            if self._cancel.is_set():
                with self._lock:
                    self._terminate_process_locked()
                return
            line = self._process.stdout.readline() if self._process.stdout else ""
            if line:
                output.append(line)
            elif self._process.poll() is not None:
                break
            else:
                time.sleep(0.05)
        exit_code = self._process.wait()
        if exit_code != 0:
            raise RuntimeError(f"{label} failed (exit {exit_code})\n{''.join(output)}")


runner = JobRunner()


class WebhookHandler(BaseHTTPRequestHandler):
    server_version = "PacketProgressWebhook/1.0"

    def do_GET(self) -> None:
        if self.path != "/health":
            self.send_error(404)
            return
        self._send_json({"status": "ok"})

    def do_POST(self) -> None:
        if self.path != "/webhook":
            self.send_error(404)
            return
        payload = self.rfile.read(int(self.headers.get("Content-Length", "0")))
        if not self._verify_signature(payload):
            return
        runner.start()
        self._send_json({"status": "queued"})

    def log_message(self, format: str, *args: object) -> None:
        return

    def _send_json(self, payload: dict) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _verify_signature(self, payload: bytes) -> bool:
        secret = os.environ.get("WEBHOOK_SECRET")
        if not secret:
            self.send_error(500, "WEBHOOK_SECRET is not set")
            return False
        signature = self.headers.get("X-Hub-Signature-256", "")
        if not signature.startswith("sha256="):
            self.send_error(401, "missing signature")
            return False
        expected = hmac.new(secret.encode("utf-8"), payload, hashlib.sha256).hexdigest()
        provided = signature.split("=", 1)[1]
        if not hmac.compare_digest(expected, provided):
            self.send_error(401, "invalid signature")
            return False
        return True


def main() -> None:
    host = os.environ.get("WEBHOOK_HOST", "0.0.0.0")
    port = int(os.environ.get("WEBHOOK_PORT", "8080"))
    server = ThreadingHTTPServer((host, port), WebhookHandler)
    server.serve_forever()


if __name__ == "__main__":
    main()
