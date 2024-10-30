package logger

import (
	"log"
	"os"
)

var (
	WarningLog *log.Logger
	InfoLog    *log.Logger
	ErrorLog   *log.Logger
	FatalLog   *log.Logger
)

func init() {
	InfoLog = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime|log.Lmsgprefix)
	WarningLog = log.New(os.Stdout, "WARNING ", log.Ldate|log.Ltime|log.Lmsgprefix)
	ErrorLog = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lmsgprefix)
	FatalLog = log.New(os.Stderr, "FATAL ", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lshortfile)
}
