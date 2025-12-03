package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 開発初期段階ではシンプルなHTTP APIでGoエンジンが動作していることを確認
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Go Audio Engine is running!")
	})

	// gRPC/WebSocketインターフェースの初期化をここに追加する

	fmt.Println("Go Audio Engine starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
