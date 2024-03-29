package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/spin-org/thermomatic/internal/device"
)

type httpd struct {
	core *core
	port uint
}
type stats struct {
	NumConnectedClients int               `json:"numConnectedClients"`
	NumCPU              int               `json:"numCpu"`
	NumGoroutine        int               `json:"numGoroutine"`
	MemStats            *runtime.MemStats `json:"memStats"`
	//TODO add bytes per second
}

type timeStampedReading struct {
	TimestampEpoch int64           `json:"timestampEpoch"`
	Reading        *device.Reading `json:"reading"`
}

func newHttpd(core *core, port uint) *httpd {
	return &httpd{
		core: core,
		port: port,
	}
}

func (d *httpd) statsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		log.Printf("[httpd] %s method not allowed ", req.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	stats := &stats{
		NumConnectedClients: d.core.numConnectedDevices(),
		NumCPU:              runtime.NumCPU(),
		NumGoroutine:        runtime.NumGoroutine(),
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	stats.MemStats = &memStats
	d.writeJSONResponse(w, *stats)
}

func imeiFromPath(path, prefix string) (uint64, error) {
	imeiStr := strings.TrimPrefix(path, prefix)
	imei, err := strconv.Atoi(imeiStr)
	if err != nil {
		log.Printf("[httpd] %s string can not be parsed as integer, %v", imeiStr, err)
		return 0, err
	}
	return uint64(imei), err
}

func (d *httpd) readingsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		log.Printf("[httpd] %s method not allowed ", req.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	imei, err := imeiFromPath(req.URL.Path, "/readings/")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lastReadingEpoch, lastReading, exists := d.core.deviceLastReading(imei)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reading := &timeStampedReading{
		TimestampEpoch: lastReadingEpoch,
		Reading:        lastReading,
	}
	d.writeJSONResponse(w, *reading)
}

func (d *httpd) statusHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		log.Printf("[httpd] %s method not allowed ", req.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	imei, err := imeiFromPath(req.URL.Path, "/status/")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, exists := d.core.deviceByIMEI(imei)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf("{\"online\":%v}", exists)))
}

func (d *httpd) run() {
	http.HandleFunc("/stats", d.statsHandler)
	http.HandleFunc("/readings/", d.readingsHandler)
	http.HandleFunc("/status/", d.statusHandler)
	httpAddress := fmt.Sprintf(":%d", d.port)
	http.ListenAndServe(httpAddress, d.logRequest(http.DefaultServeMux))
	log.Printf("[httpd] started at %s", httpAddress)
}

func (d *httpd) writeJSONResponse(w http.ResponseWriter, v interface{}) {
	json, err := json.Marshal(v)
	if err != nil {
		log.Printf("[httpd] ERR trying to serialize %v to json %v", v, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (d *httpd) logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[httpd]%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
