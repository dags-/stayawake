package wake

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

func (m *manager) startServer(ip, port string) {
	addr := fmt.Sprintf(`%s:%s`, ip, port)
	fs := http.FileServer(http.Dir("public"))
	mux := http.NewServeMux()
	mux.HandleFunc("/config", m.handleConfig)
	mux.Handle("/", fs)
	http.ListenAndServe(addr, mux)
}

func (m *manager) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg := loadCfg()

	if r.Method == "GET" {
		logger.Println("received config GET request")
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w).Encode(cfg)
		if e != nil {
			logger.Println(e)
		}
		return
	}

	if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
		logger.Println("received config POST request")
		var cfg Config
		e := json.NewDecoder(r.Body).Decode(&cfg)
		if e == nil {
			saveCfg(&cfg)
			m.setDevices(cfg.Devices)
		} else {
			logger.Println(e)
		}
		return
	}

	logger.Println("rejected", r.Method, "request from", r.RemoteAddr)
}

func ip() string {
	conn, e := net.Dial("udp", "8.8.8.8:80")
	if e != nil {
		panic(e)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func port(port string) string {
	if port == "" {
		port = "0"
	}
	add := fmt.Sprintf("127.0.0.1:%s", port)
	l, e := net.Listen("tcp", add)
	if e != nil {
		panic(e)
	}
	defer l.Close()
	parts := strings.Split(l.Addr().String(), ":")
	return parts[1]
}

func hostname(ip string) string {
	if n, e := os.Hostname(); e == nil {
		return n
	}
	return ip
}
