package main

//TODO: unittests

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bestmethod/go-logger"
	"github.com/julienschmidt/httprouter"
	"github.com/leonelquinteros/gorand"
	"net/http"
	"os"
	"strings"
)

type configLogger struct {
	LogToConsole  bool
	ErrorToStderr bool
	LogToDevlog   bool
}

type configRest struct {
	ListenIp   *string
	ListenPort int
	UseSSL     bool
	SSLCrtPath *string
	SSLKeyPath *string
}

type configRemoteLogger struct {
	EscapeCarriageReturn bool
	AddTruncated         bool
	MaxLength            int
}

type config struct {
	Logger       *configLogger
	Rest         *configRest
	RemoteLogger *configRemoteLogger
}

var logger *Logger.Logger
var rc *configRemoteLogger

// monkey patching for tests
var httpServe = http.ListenAndServe
var httpServeTls = http.ListenAndServeTLS
var psNameCall = psName
var fakeRun = false

func main() {
	//vars
	var configFile string
	var err error
	var conf config

	//parse config filename
	fmt.Printf("Parsing command line arguments...")
	flag.StringVar(&configFile, "config", "rest2log-config.txt", "Specify configuration file name and path")
	flag.Parse()
	fmt.Println("OK")

	//init basic logger
	fmt.Printf("Initializing logger...")
	logger = new(Logger.Logger)
	err = logger.Init("LOCAL", "rest2log", Logger.LEVEL_DEBUG|Logger.LEVEL_INFO|Logger.LEVEL_WARN, Logger.LEVEL_ERROR|Logger.LEVEL_CRITICAL, Logger.LEVEL_NONE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CRITICAL Could not initialize logger. Quitting. Details: %s\n", err)
		os.Exit(3)
	}
	fmt.Println("OK")

	//load configuration
	logger.Info("Loading configuration file")
	if _, err = toml.DecodeFile(configFile, &conf); err != nil {
		logger.Fatal(fmt.Sprintf("Could not parse configuration file. Quitting. Details: %s", err), 1)
	}

	//reload logger
	logger.Info("Reloading logger parameters")
	oldLogger := logger
	logger = new(Logger.Logger)
	var devlog int
	var stdout int
	var stderr int
	if conf.Logger.LogToDevlog == true {
		devlog = Logger.LEVEL_DEBUG | Logger.LEVEL_INFO | Logger.LEVEL_WARN | Logger.LEVEL_ERROR | Logger.LEVEL_CRITICAL
	} else {
		devlog = Logger.LEVEL_NONE
	}
	if conf.Logger.LogToConsole == true {
		stdout = Logger.LEVEL_DEBUG | Logger.LEVEL_INFO | Logger.LEVEL_WARN
	}
	if conf.Logger.ErrorToStderr == true {
		stderr = Logger.LEVEL_ERROR | Logger.LEVEL_CRITICAL
	} else {
		stderr = Logger.LEVEL_NONE
		stdout = stdout | Logger.LEVEL_ERROR | Logger.LEVEL_CRITICAL
	}
	err = logger.Init("REMOTE", "rest2log", stdout, stderr, devlog)
	if err != nil {
		oldLogger.Fatal(fmt.Sprintf("Could not reload logger from configuration, devlog issue? Quitting. Details: %s", err), 4)
	}

	rc = conf.RemoteLogger
	//start listener
	router := httprouter.New()
	router.GET("/", help)
	router.POST("/:logLevel", logLine)
	oldLogger.Info(fmt.Sprintf("Listening on %s:%d TLS=%t", *conf.Rest.ListenIp, conf.Rest.ListenPort, conf.Rest.UseSSL))
	if conf.Rest.UseSSL == false {
		err = httpServe(fmt.Sprintf("%s:%d", *conf.Rest.ListenIp, conf.Rest.ListenPort), router)
	} else {
		err = httpServeTls(fmt.Sprintf("%s:%d", *conf.Rest.ListenIp, conf.Rest.ListenPort), *conf.Rest.SSLCrtPath, *conf.Rest.SSLKeyPath, router)
	}
	if err != nil {
		oldLogger.Fatal(fmt.Sprintf("Could not listen and serve. Quitting. Details: %s", err), 9)
	}
}

func help(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("CALL: /{LEVEL}\nBODY: POST: {\"message\":\"Your Message To Log Here\"}\n"))
}

type LogBody struct {
	Message string
}

func logLine(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body == nil {
		http.Error(w, "Request body missing!", 400)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var b LogBody
	err := decoder.Decode(&b)
	if err != nil {
		http.Error(w, "Invalid json provided in post content. Must be: {\"message\":\"Your Message To Log Here\"}", 400)
		return
	}
	defer r.Body.Close()
	b.Message = strings.Replace(b.Message, "\r", "", -1)
	if rc.EscapeCarriageReturn == true {
		b.Message = strings.Replace(b.Message, "\n", "\\n", -1)
	}
	var i int
	if len(b.Message) > rc.MaxLength {
		if rc.AddTruncated == false {
			b.Message = b.Message[0:rc.MaxLength]
		} else {
			i = rc.MaxLength - 11
			b.Message = b.Message[0:i]
			b.Message = strings.Join([]string{b.Message, "(truncated)"}, "")
		}
	}

	var Messages []string
	var uuid string
	if rc.EscapeCarriageReturn == false {
		Messages = strings.Split(b.Message, "\n")
		uuidb, _ := gorand.UUIDv4()
		uuid, _ = gorand.MarshalUUID(uuidb)
	} else {
		Messages = append(Messages, b.Message)
	}

	for i := range Messages {

		switch strings.ToUpper(psNameCall(ps)) {
		case "DEBUG":
			if len(Messages) > 1 {
				logger.Debug(fmt.Sprintf("multipart_id=%s %s", uuid, Messages[i]))
			} else {
				logger.Debug(Messages[i])
			}
		case "INFO":
			if len(Messages) > 1 {
				logger.Info(fmt.Sprintf("multipart_id=%s %s", uuid, Messages[i]))
			} else {
				logger.Info(Messages[i])
			}
		case "WARN":
			if len(Messages) > 1 {
				logger.Warn(fmt.Sprintf("multipart_id=%s %s", uuid, Messages[i]))
			} else {
				logger.Warn(Messages[i])
			}
		case "ERROR":
			if len(Messages) > 1 {
				logger.Error(fmt.Sprintf("multipart_id=%s %s", uuid, Messages[i]))
			} else {
				logger.Error(Messages[i])
			}
		case "CRITICAL":
			if len(Messages) > 1 {
				logger.Critical(fmt.Sprintf("multipart_id=%s %s", uuid, Messages[i]))
			} else {
				logger.Critical(Messages[i])
			}
		default:
			http.Error(w, "Invalid logger type. Must be one of: DEBUG | INFO | WARN | ERROR | CRITICAL", 400)
			return
		}
	}
	if fakeRun == false {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	}
}

func psName(ps httprouter.Params) string {
	return ps.ByName("logLevel")
}
