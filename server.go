package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	r := mux.NewRouter()

	// 사이트 홈을 public/index.html로 지정
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/index.html")
	})

	r.HandleFunc("/ws", handleWebSocket)
	r.HandleFunc("/video/{filename}", handleVideoStream)

	log.Println("Server started on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// 클라이언트에서 WebSocket 연결이 성공적으로 열리면 해당 로직이 실행됩니다.
	log.Println("WebSocket connection established")

	// 동영상 파일을 열어서 전송할 준비를 합니다.
	file, err := os.Open("path/to/video.mp4")
	if err != nil {
		log.Println("Failed to open video file:", err)
		return
	}
	defer file.Close()

	bufferSize := 1024 * 4 // 4KB의 버퍼 크기로 설정합니다.
	buffer := make([]byte, bufferSize)

	// 파일을 읽어서 WebSocket을 통해 클라이언트로 전송합니다.
	for {
		// 버퍼에 데이터를 읽어옵니다.
		n, err := file.Read(buffer)
		if err != nil {
			log.Println("Failed to read video file:", err)
			break
		}

		// 버퍼의 내용을 클라이언트로 전송합니다.
		if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
			log.Println("Failed to send video data over WebSocket:", err)
			break
		}

		// 전송 속도 제어를 위해 10ms 대기합니다.
		time.Sleep(10 * time.Millisecond)
	}

	log.Println("Video streaming completed")
}

func handleVideoStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	filePath := filepath.Join(".", "videos", filename)

	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Failed to open video file:", err)
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "video/mp4")
	http.ServeContent(w, r, "", time.Now(), file)
}
