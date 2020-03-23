package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sccxx/go-mjpeg"
)

var (
	addr = ":8080"
)

func capture(ctx context.Context, wg *sync.WaitGroup, stream *mjpeg.Stream) {
	defer wg.Done()

	for len(ctx.Done()) == 0 {
		var data []byte
		if stream.NWatch() > 0 {
			img, err := CaptureScreen()
			if err != nil {
				continue
			}
			if err != nil {
				continue
			}
			buf := new(bytes.Buffer)
			err = jpeg.Encode(buf, img, nil)
			if err != nil {
				continue
			}
			data = buf.Bytes()
		}
		err := stream.Update(data)
		if err != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	stream := mjpeg.NewStream()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go capture(ctx, &wg, stream)

	http.HandleFunc("/mjpeg", stream.ServeHTTP)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<img src="/mjpeg" />`))
	})

	server := &http.Server{Addr: addr}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		<-sc
		server.Shutdown(ctx)
	}()
	server.ListenAndServe()
	stream.Close()
	cancel()

	wg.Wait()
	fmt.Println()
	fmt.Println("bye")
}
