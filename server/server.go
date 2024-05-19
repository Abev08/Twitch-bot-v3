package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"twitch_bot_v3/database"
)

var ServerAddress string = "127.0.0.1:3000"

var initialized bool
var notificationsHtml, configHtml string

func Init() {
	if initialized {
		return
	}
	initialized = true

	slog.Info(
		"Starting HTTP server",
		"obs", fmt.Sprintf("http://%s", ServerAddress),
		"dashboard", fmt.Sprintf("http://%s/dashboard", ServerAddress),
		"config", fmt.Sprintf("http://%s/config", ServerAddress))

	// Load server files
	if data, err := os.ReadFile("server/notifications.html"); err != nil {
		slog.Error("Error when reading notifications.html", "err", err)
		return
	} else {
		notificationsHtml = string(data)
	}
	if data, err := os.ReadFile("server/config.html"); err != nil {
		slog.Error("Error when reading config.html", "err", err)
		return
	} else {
		configHtml = string(data)
	}

	// Start server loop
	go update()

	// If minimum required data is missing open browser to force user to provide it
	// if !database.IsRequiredInfoProvided() {
	// 	if err := OpenUrl(fmt.Sprintf("http://%s/config", SERVER_ADDRESS)); err != nil {
	// 		slog.Error("Couldn't open browser window", "err", err)
	// 	}
	// }
}

func update() {
	var tempBuff = make([]byte, 65535)
	listener, err := net.Listen("tcp", ServerAddress)
	if err != nil {
		slog.Error("HTTP server error", "err", err)
		return
	}

	for {
		var conn net.Conn
		var err error
		var req *http.Request
		var resp *http.Response
		if conn, err = listener.Accept(); err != nil {
			slog.Warn("HTTP server, error on new connection", "err", err)
			conn.Close()
			continue
		}

		n, err := conn.Read(tempBuff)
		if err != nil {
			slog.Error("HTTP server, connection read error", "err", err)
		} else {
			var reader = bufio.NewReader(bytes.NewReader(tempBuff[:n]))
			req, err = http.ReadRequest(reader)
			if err != nil {
				slog.Error("HTTP server, connection request parse error", "err", err)
				conn.Close()
				continue
			}
			// fmt.Print(req)
			// sb.WriteString(fmt.Sprintf("New http %s request, url: %s", req.Method, req.URL))
			// for _, v := range req.Header.Values("Upgrade") {
			// 	sb.WriteString(fmt.Sprintf(", requested upgrade: %s", v))
			// }
			slog.Info("HTTP server, new request", "url", req.URL)

			resp = &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Request:    req,
				Header:     http.Header{},
			}

			switch req.URL.String() {
			case "/":
				resp.Body = io.NopCloser(strings.NewReader(notificationsHtml))
				resp.ContentLength = int64(len(notificationsHtml))
			case "/config":
				resp.Body = io.NopCloser(strings.NewReader(configHtml))
				resp.ContentLength = int64(len(configHtml))
			case "/config_data":
				d := database.GetSecretsAndConfigAsJson()
				resp.Body = io.NopCloser(bytes.NewReader(d))
				resp.ContentLength = int64(len(d))
			case "/secrets_update":
				var responseMap []string
				err = json.Unmarshal([]byte(req.Header.Values("Secrets")[0]), &responseMap)
				if err != nil {
					slog.Error("Couldn't parse secrets update request", "err", err)
				} else {
					for i, v := range responseMap {
						switch i {
						case 0:
							database.UpdateSecretsValue(database.TwitchName, v)
						case 1:
							database.UpdateSecretsValue(database.TwitchCustomerID, v)
						case 2:
							database.UpdateSecretsValue(database.TwitchPassword, v)
						}
					}
				}
			case "/config_update":
				var responseMap []string
				err = json.Unmarshal([]byte(req.Header.Values("Config")[0]), &responseMap)
				if err != nil {
					slog.Error("Couldn't parse secrets update request", "err", err)
				} else {
					for i, v := range responseMap {
						switch i {
						case 0:
							database.UpdateConfigValue(database.ChannelName, v)
						}
					}
				}
			default:
				resp.Status = "404 Not Found"
				resp.StatusCode = 404
			}
		}

		if resp != nil {
			resp.Write(conn)
		}
		conn.Close()
	}
}

func OpenUrl(url string) error {
	var err error = nil

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}
