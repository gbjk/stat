// Copyright 2011 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//      Unless required by applicable law or agreed to in writing, software
//      distributed under the License is distributed on an "AS IS" BASIS,
//      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//      See the License for the specific language governing permissions and
//      limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/gbjk/stat"
)

var (
	listenAddr = flag.String("http", ":8090", "HTTP listen port")
	maxLen     = flag.Int("max", 60, "max points to retain")
)

type Server struct {
	series map[string][][2]int64
	start  time.Time
	mu     sync.Mutex
}

func NewServer() *Server {
	return &Server{
		series: make(map[string][][2]int64),
		start:  time.Now(),
	}
}

func (s *Server) Update(args *stat.Point, r *struct{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// append point to series
	key := args.Process + " " + args.Series
	second := int64(time.Now().Sub(s.start)) / 100e6
	s.series[key] = append(s.series[key], [2]int64{second, args.Value})
	// trim series to maxLen
	if sk := s.series[key]; len(sk) > *maxLen {
		s.series[key] = sk[len(sk)-*maxLen:]
	}
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(w)
	e.Encode(s.series)
}

func Static(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[1:]
	if filename == "" {
		filename = "index.html"
	} else if filename[:6] != "flotr/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/"+filename)
}

func main() {
	flag.Parse()
	server := NewServer()
	rpc.Register(server)
	rpc.HandleHTTP()
	http.HandleFunc("/", Static)
	http.Handle("/get", server)

	fmt.Println("Serving up stats on ", *listenAddr)
	http.ListenAndServe(*listenAddr, nil)
}
