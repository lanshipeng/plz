package countlog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type fileLogOutput struct {
	windowSize          int64
	logFile             string
	rotateAfter         int64
	openedFile          *os.File
	openedFileArchiveTo string
}

func (output *fileLogOutput) Close() {
	if output.openedFile != nil {
		output.openedFile.Close()
	}
}

func (output *fileLogOutput) openLogFile() {
	var err error
	output.openedFile, err = os.OpenFile(output.logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		os.Stderr.Write([]byte("failed to open log file: " +
			output.logFile + ", " + err.Error() + "\n"))
		os.Stderr.Sync()
	}
	output.openedFileArchiveTo = output.logFile + "." + time.Now().Format("200601021504")
}

func (output *fileLogOutput) archiveLogFile() {
	output.openedFile.Close()
	output.openedFile = nil
	err := os.Rename(output.logFile, output.openedFileArchiveTo)
	if err != nil {
		os.Stderr.Write([]byte("failed to rename to archived log file: " +
			output.openedFileArchiveTo + ", " + err.Error() + "\n"))
		os.Stderr.Sync()
	}
}

func (output *fileLogOutput) OutputLog(level int, timestamp int64, formattedEvent []byte) {
	if timestamp > output.rotateAfter {
		now := time.Now()
		output.rotateAfter = (int64(now.UnixNano()/output.windowSize) + 1) * output.windowSize
		output.archiveLogFile()
		output.openLogFile()
	}
	if output.openedFile != nil {
		output.openedFile.Write(formattedEvent) // silently ignore error
	}
}

type osFileLogOutput struct {
	logFile *os.File
}

func (output *osFileLogOutput) Close() {
	output.logFile.Sync()
}

func (output *osFileLogOutput) OutputLog(level int, timestamp int64, formattedEvent []byte) {
	output.logFile.Write(withColorLevelPrefix(level, formattedEvent))
}

func NewFileLogOutput(logFile string) LogOutput {
	switch logFile {
	case "STDOUT":
		return &osFileLogOutput{os.Stdout}
	case "STDERR":
		return &osFileLogOutput{os.Stderr}
	default:
		output := &fileLogOutput{
			windowSize: int64(time.Hour),
			logFile:    logFile,
		}
		err := os.MkdirAll(filepath.Dir(logFile), 0755)
		if err != nil {
			os.Stderr.Write([]byte("failed to create dir for log file: " +
				filepath.Dir(logFile) + ", " + err.Error() + "\n"))
			os.Stderr.Sync()
		}
		output.openLogFile()
		output.rotateAfter = (int64(time.Now().UnixNano()/output.windowSize) + 1) * output.windowSize
		return output
	}
}

func getLevelName(level int) string {
	switch level {
	case LevelTrace: return "TRACE"
	case LevelDebug: return "DEBUG"
	case LevelInfo: return "INFO"
	case LevelWarn: return "WARN"
	case LevelError: return "ERROR"
	case LevelFatal: return "FATAL"
	default: return "UNKNOWN"
	}
}

func withColorLevelPrefix(level int, msg []byte) []byte {
	levelColor := getColor(level)
	// ESC = \x1b
	// ESC+[ =  Control Sequence Introducer
	// \x1b[%d;1m, eg. \x1b32;1m
	// \x1b[0m
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "\x1b[%d;1m[%s]\x1b[0m%s", levelColor, getLevelName(level), msg)
	return buf.Bytes()
}

const (
	nocolor = 0
	black   = 30
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	purple  = 35
	cyan    = 36
	gray    = 37
)

func getColor(level int) int {
	switch level {
	case LevelTrace: return cyan
	case LevelDebug: return gray
	case LevelInfo: return green
	case LevelWarn: return yellow
	case LevelError: return red
	case LevelFatal: return purple
	default: return nocolor
	}
}
