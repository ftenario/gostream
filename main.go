package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/go-zoo/bone"
)

const size = 1024 * 8

func main() {
	fmt.Println("Hello")

	mux := bone.New()
	mux.Get("/", http.HandlerFunc(home))
	mux.Get("/movie.mp4", http.HandlerFunc(stream))

	n := negroni.Classic()
	n.UseHandler(mux)

	n.Run(":3000")
}

func home(w http.ResponseWriter, req *http.Request) {
	// if req.URL.Path != "/index.html" {
	// 	http.Error(w, "Method not found", 404)
	// 	return
	// }

	if req.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	f, err := os.Open("index.html")
	if err == nil {
		defer f.Close()
		w.Header().Add("Content-Type", "text/html")
		br := bufio.NewReader(f)
		br.WriteTo(w)
	} else {
		fmt.Println(err)
		w.WriteHeader(404)
	}

}

func stream(w http.ResponseWriter, req *http.Request) {

	file, err := os.Open("movie.mp4")
	if err != nil {
		w.WriteHeader(500)
		return
	}

	defer file.Close()

	f, err := file.Stat()
	if err != nil {
		w.WriteHeader(500)
	}

	fileSize := int(f.Size())
	fmt.Println("File Size: ", fileSize)
	total := strconv.Itoa(fileSize)
	iTotal, _ := strconv.Atoi(total)

	//rangeValue := strings.Split(req.Header.Get("Range"), "=")[1]
	rangeValue := strings.Split(req.Header.Get("Range"), "=")

	fmt.Println("range Value: ", rangeValue)
	var start string
	var params []string
	var end string

	if len(rangeValue) < 2 {
		start = "0"
		end = strconv.Itoa(iTotal - 1)
	} else {
		params = strings.Split(rangeValue[1], "-")
		start = params[0]
		if params[1] != "" {
			end = params[1]
		} else {
			end = strconv.Itoa(iTotal - 1)
		}
	}

	// fmt.Println("Params len: ", len(params))

	iEnd, _ := strconv.Atoi(end)
	iStart, _ := strconv.Atoi(start)
	chunkSize := (iEnd - iStart) + 1
	// fmt.Println("Start: ", start)
	// fmt.Println("End: ", end)
	// fmt.Println("ChunkSize: ", strconv.Itoa(chunkSize))

	w.Header().Set("Content-Range", "bytes "+start+"-"+end+"/"+total)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", strconv.Itoa(chunkSize))
	w.WriteHeader(206)

	buffer := make([]byte, size)
	s, _ := strconv.Atoi(start)
	file.Seek(int64(s), 0)
	writeBytes := 0

	for {
		n, err := file.Read(buffer)
		writeBytes += n

		if err != nil {
			break
		}

		data := buffer[:n]
		w.Write(data)
		w.(http.Flusher).Flush()
	}
}
