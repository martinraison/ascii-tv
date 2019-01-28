package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type frame struct {
	lines    []string
	duration time.Duration
}

var speed = 1 // Default speed of the movie

func main() {
	moviePtr := flag.String("movie", "resources/sw1.txt", "path to ASCII movie file")
	addrPtr := flag.String("addr", ":8080", "TCP address to listen on")
	flag.Parse()
	data, err := ioutil.ReadFile(*moviePtr)
	if err != nil {
		fmt.Printf("Failed to load file %s\n", *moviePtr)
	}
	lines := strings.Split(string(data), "\n")
	frameHeight := 13
	var frames []frame
	for i := range lines {
		if i%(frameHeight+1) == 0 {
			frameDurationStr := lines[i]
			frameDurationInt, err := strconv.ParseInt(frameDurationStr, 0, 64)
			if err != nil {
				fmt.Printf("Failed to parse frame duration from line: %s", frameDurationStr)
			}
			frames = append(frames, frame{lines[i+1 : i+1+frameHeight], time.Duration(frameDurationInt)})
		}
	}
	fmt.Printf("Extracted %d frames from %s\n", len(frames), *moviePtr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(strings.ToLower(r.Header.Get("User-Agent")), "curl") {
			fmt.Fprintf(w, "Web browsers are so %d, use \"curl https://asciitv.fr\" from your terminal instead!\n", time.Now().Year()-1)
			return
		}

		if value, ok := r.URL.Query()["speed"]; ok {
			if speedInt, err := strconv.Atoi(value[0]); err == nil {
				speed = speedInt
			}
		}

		for _, frame := range frames {
			// Clear terminal and move cursor to position (1,1)
			fmt.Fprint(w, "\033[2J\033[1;1H")
			for _, line := range frame.lines {
				fmt.Fprintln(w, line)
			}

			if speed != 1 {
				fmt.Fprintln(w, "\n\nSpeed:", speed)
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep((frame.duration / time.Duration(speed)) * time.Second / 15)
		}
	})

	fmt.Printf("Listening on %s\n", *addrPtr)
	log.Fatal(http.ListenAndServe(*addrPtr, nil))
}
