package debug

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/mattn/go-tty"
)

var handlers = map[string]func(){
	"?": help,
	"c": clear,
	"l": listDevices,
	"s": toggleLogcat,
	"f": flushLogcat,
	"d": dumpLogcat,
	"+": increaseSleep,
	"-": decreaseSleep,
	"=": printDelay,
	"r": restartGame,
}

// a list of needles to search for in the process list
// to find Azur Lane's PID -- these should be lowercase
// to make the search case-insensitive
const grepRegex = "'(azurlane|blhx|manjuu|yostar)'"

// Filter to remove Azur Lane's uninteresting logs (FacebookSDK, ...) -- regex for -e parameter
// see https://developer.android.com/studio/command-line/logcat#filteringOutput
const defaultLogcatFilter = "(System|Unity)"

var logcatProcess *exec.Cmd
var azurLanePID int
var psDelay time.Duration = 3 * time.Second

const azurLanePackage = "com.YoStarEN.AzurLane" // TODO: handle other regions (CN, JP, KR, TW)

func restartGame() {
	logger.LogEvent("ADB", "Restart", "Restarting game...", logger.LOG_LEVEL_INFO)
	cmd := exec.Command("adb", "shell", "am", "force-stop", azurLanePackage)
	if err := cmd.Run(); err != nil {
		logger.LogEvent("ADB", "Restart", fmt.Sprintf("Failed to force-stop: %v", err), logger.LOG_LEVEL_ERROR)
		return
	}
	time.Sleep(3 * time.Second)
	cmd = exec.Command("adb", "shell", "monkey", "-p", azurLanePackage, "-c", "android.intent.category.LAUNCHER", "1")
	if err := cmd.Run(); err != nil {
		logger.LogEvent("ADB", "Restart", fmt.Sprintf("Failed to launch: %v", err), logger.LOG_LEVEL_ERROR)
		return
	}
	logger.LogEvent("ADB", "Restart", "Game restarted successfully", logger.LOG_LEVEL_INFO)
}

func help() {
	fmt.Println("belfast -- adb watcher help")
	fmt.Println("?: print this help")
	fmt.Println("l: list connected devices")
	fmt.Println("c: clear terminal")
	fmt.Println("s: start/stop logcat parsing")
	fmt.Println("f: flush logcat")
	fmt.Println("d: dump logcat buffer to a file")
	fmt.Println("r: restart game")
	fmt.Println("+: increase delay between ps commands (default: 3s)")
	fmt.Println("-: decrease delay between ps commands (default: 3s)")
	fmt.Println("=: print current delay between ps commands")
	fmt.Println("x: exit adb watcher")
}

// stupid way to clear the terminal, calls 'clear' on non-windows and 'cls' on windows
func clear() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// runs 'adb devices' and prints the output
func listDevices() {
	cmd := exec.Command("adb", "devices")
	out, err := cmd.Output()
	if err != nil {
		logger.LogEvent("ADB", "ListDevices", "Failed to list devices", logger.LOG_LEVEL_ERROR)
		return
	}
	fmt.Print(string(out))
}

// runs 'adb logcat -c' to flush logcat's buffer
func flushLogcat() {
	cmd := exec.Command("adb", "logcat", "-c")
	if err := cmd.Run(); err != nil {
		logger.LogEvent("ADB", "Flush", "Failed to flush logcat", logger.LOG_LEVEL_ERROR)
		return
	}
	logger.LogEvent("ADB", "FlushLogcat", "Logcat flushed", logger.LOG_LEVEL_INFO)
}

// wrapper function to print logcat lines
func echoLog(line *string) {
	fmt.Println(*line)
}

func stopLogcat() {
	if logcatProcess == nil {
		return
	}
	pid := logcatProcess.Process.Pid
	logcatProcess.Process.Kill()
	logcatProcess = nil
	logger.LogEvent("ADB", "Logcat", fmt.Sprintf("Logcat stopped (PID: %d)", pid), logger.LOG_LEVEL_INFO)
}

// starts/stops logcat parsing
func toggleLogcat() {
	if logcatProcess != nil {
		stopLogcat()
		return
	}
	go func() {
		args := []string{"logcat"}
		if azurLanePID != 0 {
			args = append(args, "--pid", fmt.Sprintf("%d", azurLanePID), "-e", defaultLogcatFilter)
		} else {
			logger.LogEvent("ADB", "Logcat", fmt.Sprintf("Azur Lane PID not found, waiting %v to retry", psDelay), logger.LOG_LEVEL_INFO)
			return
		}
		logcatProcess = exec.Command("adb", args...)
		processStdout, err := logcatProcess.StdoutPipe()
		if err != nil {
			logger.LogEvent("ADB", "Logcat", fmt.Sprintf("Failed to get logcat stdout: %v", err), logger.LOG_LEVEL_ERROR)
			return
		}

		if err := logcatProcess.Start(); err != nil {
			logger.LogEvent("ADB", "Logcat", fmt.Sprintf("Failed to start logcat: %v", err), logger.LOG_LEVEL_ERROR)
			return
		}
		logger.LogEvent("ADB", "Logcat", fmt.Sprintf("Logcat started (PID: %d)", logcatProcess.Process.Pid), logger.LOG_LEVEL_INFO)

		// Parse logcat output in background
		go func() {
			scanner := bufio.NewScanner(processStdout)
			for scanner.Scan() {
				line := scanner.Text()
				echoLog(&line)
			}
			if err := scanner.Err(); err != nil {
				logger.LogEvent("ADB", "Logcat", fmt.Sprintf("Error reading logcat stdout: %v", err), logger.LOG_LEVEL_INFO)
			}
		}()
		logcatProcess.Wait()
		exitCode := logcatProcess.ProcessState.ExitCode()
		if exitCode != 0 {
			logger.LogEvent("ADB", "Logcat", fmt.Sprintf("Logcat process (PID: %d) exited with code %d", logcatProcess.Process.Pid, exitCode), logger.LOG_LEVEL_ERROR)
		}
	}()
}

// increases by 1s the delay between ps commands
func increaseSleep() {
	psDelay += 1 * time.Second
	logger.LogEvent("Watcher", "Delay", fmt.Sprintf("Delay increased to %v", psDelay), logger.LOG_LEVEL_INFO)
}

// decreases by 1s the delay between ps commands
func decreaseSleep() {
	if psDelay > 1*time.Second {
		psDelay -= 1 * time.Second
		logger.LogEvent("Watcher", "Delay", fmt.Sprintf("Delay decreased to %v", psDelay), logger.LOG_LEVEL_INFO)
	} else {
		logger.LogEvent("Watcher", "Delay", "Delay cannot be decreased further, minimum is 1s", logger.LOG_LEVEL_INFO)
	}
}

// prints the current delay between ps commands
func printDelay() {
	logger.LogEvent("ADB", "PrintDelay", fmt.Sprintf("Current delay: %v", psDelay), logger.LOG_LEVEL_INFO)
}

// dump logcat buffer to a file
func dumpLogcat() {
	cmd := exec.Command("adb", "logcat", "-d")
	filename := time.Now().Format("2006-01-02_15-04-05") + "_belfast_logcat.log"
	file, err := os.Create(filename)
	if err != nil {
		logger.LogEvent("ADB", "DumpLogcat", "Failed to create file", logger.LOG_LEVEL_ERROR)
		return
	}
	cmd.Stdout = file
	cmd.Stderr = file
	if err := cmd.Run(); err != nil {
		logger.LogEvent("ADB", "DumpLogcat", "Failed to dump logcat", logger.LOG_LEVEL_ERROR)
		return
	}
	defer file.Close()
	logger.LogEvent("ADB", "DumpLogcat", fmt.Sprintf("Logcat dumped to %s", filename), logger.LOG_LEVEL_INFO)
}

// main routine for ADB watcher, listens for key presses and executes commands
func ADBRoutine(tty *tty.TTY, flush bool, restart bool) {
	if tty == nil {
		return // silently return, main function will handle the error
	}
	// Checking if adb is installed / in PATH
	_, err := exec.LookPath("adb")
	if err != nil {
		logger.LogEvent("ADB", "Init", "ADB not found in PATH", logger.LOG_LEVEL_ERROR)
		return
	}
	logger.LogEvent("ADB", "Init", "ADB watcher started", logger.LOG_LEVEL_INFO)
	if flush {
		flushLogcat()
	}
	if restart {
		restartGame()
	}
	help()

	// Goroutine to find Azur Lane's PID
	go func() {
		for {
			cmd := exec.Command("adb", "shell", "ps", "-A", "-o", "PID,NAME", "|", "grep", "-iE", grepRegex)
			if out, err := cmd.Output(); err == nil {
				for _, line := range strings.Split(string(out), "\n") {
					newPid := 0
					fmt.Sscanf(line, "%d", &newPid)
					if newPid != 0 && newPid != azurLanePID {
						azurLanePID = newPid
						logger.LogEvent("ADB", "Shell", fmt.Sprintf("Azur Lane PID: %d", azurLanePID), logger.LOG_LEVEL_INFO)
						stopLogcat()   // force stop logcat to restart with new PID
						toggleLogcat() // restart logcat with new PID
					}
				}
			}
			time.Sleep(psDelay)
		}
	}()

	for {
		// Read key from stdin
		// If key is '?' print hello world
		r, err := tty.ReadRune()
		if err != nil {
			logger.LogEvent("ADB", "ReadRune", fmt.Sprintf("Failed to read rune: %v", err), logger.LOG_LEVEL_ERROR)
			break
		}
		if r == 'x' {
			logger.LogEvent("ADB", "Exit", "ADB watcher exited", logger.LOG_LEVEL_INFO)
			break
		}
		if handler, ok := handlers[string(r)]; ok {
			handler()
		}
	}
}
