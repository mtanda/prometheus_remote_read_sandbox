package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

func remoteReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}

	compressed, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req prompb.ReadRequest
	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		http.Error(w, "Unable to parse request", http.StatusBadRequest)
		return
	}
	fmt.Printf("%+v\n", req)

	ts := &prompb.TimeSeries{}
	resp := prompb.ReadResponse{
		Results: []*prompb.QueryResult{
			{Timeseries: []*prompb.TimeSeries{ts}},
		},
	}

	respBuf, err := proto.Marshal(&resp)
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(snappy.Encode(nil, respBuf)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/read", remoteReadHandler)
	log.Println("Starting server on :9415")
	if err := http.ListenAndServe(":9415", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
