package proxychangerlib

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/gotk3/gotk3/glib"
	"github.com/juju/loggo"
)

const DBUS_PATH = "/com/github/okelet/proxychanger"
const DBUS_INTERFACE = "com.github.okelet.proxychanger"

var DEFAULT_EXCLUDED_INTERFACES_REGEXPS []string

const DEFAULT_TIME_BETWEEN_IP_CHECKS = 10
const DEFAULT_TIME_BETWEEN_UPDATE_CHECKS = 1800

const APP_ID = "proxychanger"
const ICON_NAME = "proxychanger"
const GETTEXT_DOMAIN = "proxychanger"

var HOME_DIR string
var APP_DIR string
var LOCALE_DIR string
var AUTOSTART_DIR string
var AUTOSTART_FILE string

const LOG_FILENAME = "proxychanger.log"

var LOG_PATH string

const DEFAULT_CONFIG_FILE = "proxychanger.json"

var DEFAULT_CONFIG_PATH string

var Log loggo.Logger

// Replace the default formatter with a new one that respects time zone
// https://github.com/juju/loggo/blob/master/formatter.go
func TzLoggoFormatter(entry loggo.Entry) string {
	ts := entry.Timestamp.Format("2006-01-02 15:04:05")
	filename := filepath.Base(entry.Filename)
	return fmt.Sprintf("%s %s %s %s:%d %s", ts, entry.Level, entry.Module, filename, entry.Line, entry.Message)
}

func InitConstants() error {

	DEFAULT_EXCLUDED_INTERFACES_REGEXPS = []string{"^lo$", "^virbr[0-9]+$", "virbr[0-9]+-nic", "docker[0-9]+"}

	HOME_DIR = glib.GetHomeDir()
	if HOME_DIR == "" {
		return errors.New("Empty user home dir")
	}

	APP_DIR = path.Join(HOME_DIR, ".proxychanger")

	DEFAULT_CONFIG_PATH = path.Join(APP_DIR, DEFAULT_CONFIG_FILE)
	LOG_PATH = path.Join(APP_DIR, LOG_FILENAME)

	LOCALE_DIR = path.Join(APP_DIR, "locale")
	AUTOSTART_DIR = path.Join(glib.GetUserConfigDir(), "autostart")
	AUTOSTART_FILE = path.Join(AUTOSTART_DIR, "proxychanger.desktop")

	Log = loggo.GetLogger("com.github.okelet.proxychanger")

	fileRotateWriter := &lumberjack.Logger{
		Filename:   LOG_PATH,
		MaxSize:    5,
		MaxBackups: 7,
	}

	// Always log to file (removing the default standard output)
	Log.SetLogLevel(loggo.WARNING)
	loggo.ReplaceDefaultWriter(loggo.NewSimpleWriter(fileRotateWriter, func(entry loggo.Entry) string {
		return TzLoggoFormatter(entry)
	}))

	return nil

}

// Add a new writer to log to the error output
func AddErrorOutputLogging() {
	loggo.RegisterWriter("stderr", loggo.NewSimpleWriter(os.Stderr, func(entry loggo.Entry) string {
		return TzLoggoFormatter(entry)
	}))
}
