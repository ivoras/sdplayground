package main

import (
	"database/sql"
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

const (
	eventQuit = iota
)

type sysEventMessage struct {
	event int
	idata int
}

var sysEventChannel = make(chan sysEventMessage, 5)
var logOutput io.Writer
var startTime time.Time

var logFileName = flag.String("log", "/tmp/af.log", "Log file ('-' for only stderr)")

var db *bun.DB

func main() {
	os.Setenv("TZ", "UTC")
	startTime = time.Now()
	rand.Seed(startTime.UnixNano())

	if runtime.GOOS == "windows" {
		*logFileName = "c:\\temp\\af.log"
	}
	flag.Parse()

	if *logFileName != "-" {
		f, err := os.OpenFile(*logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.Panic("Cannot open log file " + *logFileName)
		}
		defer f.Close()
		logOutput = io.MultiWriter(os.Stderr, f)
	} else {
		logOutput = os.Stderr
	}
	log.SetOutput(logOutput)

	log.Println("Starting up...")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(os.Getenv("DB_DSN"))))
	db = bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook())

	defer db.Close()
	defer sqldb.Close()

	go webServer()
	//go infraWebServer()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	oldAlloc := int64(m.Alloc)
	printMemStats(&m)

	for {
		select {
		case msg := <-sysEventChannel:
			switch msg.event {
			case eventQuit:
				log.Println("Exiting")
				os.Exit(msg.idata)
			}
		case sig := <-sigChannel:
			switch sig {
			case syscall.SIGINT:
				sysEventChannel <- sysEventMessage{event: eventQuit, idata: 0}
				log.Println("^C detected")
			}
		case <-time.After(60 * time.Second):

			runtime.ReadMemStats(&m)
			if abs(int64(m.Alloc)-oldAlloc) > 1024*1024 {
				printMemStats(&m)
				oldAlloc = int64(m.Alloc)
			}
		case <-time.After(15 * time.Minute):
			//cleanupDb()
		}
	}
}

func printMemStats(m *runtime.MemStats) {
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.Printf("Alloc: %v MiB\tTotalAlloc: %v MiB\tSys: %v MiB\tNumGC: %v\tUptime: %0.1fh\n",
		bToMB(m.Alloc), bToMB(m.TotalAlloc), bToMB(m.Sys), m.NumGC, time.Since(startTime).Hours())
}
