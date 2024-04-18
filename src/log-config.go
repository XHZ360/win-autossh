package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type MyLogWriter struct {
	lastTime time.Time
	logfile  *os.File
}

const logDirName = "./logs"

var logDir = ""

func sameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}
func (m MyLogWriter) Write(p []byte) (n int, err error) {
	now := time.Now()
	// check if need to create new log file
	if sameDay(now, m.lastTime) {
		if m.logfile != nil {
			err := m.logfile.Close()
			if err != nil {
				panic("close log file repeatedly")
			}
			m.logfile = nil
		}
	}
	m.lastTime = now
	// check if need to open new log file
	if m.logfile == nil {
		filename := now.Format("2006-01-02") + ".log"
		path := logDir + "/" + filename
		logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to open log file")
		}
		m.logfile = logFile
	}
	// write log and trim date
	return m.logfile.Write(p[11:])
}

/**
 * logConfig configures the log output
 */
func logConfig(workDir string) {
	logDir = filepath.Join(workDir, logDirName)
	// check log directory exists or not
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, 0755)
		if err != nil {
			log.Fatal("Failed to create log directory: ", err)
		}
	}

	writer := MyLogWriter{lastTime: time.Now()}
	log.SetOutput(writer)
}
