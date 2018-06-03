package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gocraft/web"
	"github.com/mlctrez/gflamescope/gfutil"
	"github.com/mlctrez/gflamescope/heatmap"
	"github.com/mlctrez/gflamescope/stack"
	"github.com/mlctrez/zipbackpack/httpfs"
)

var root string

type Context struct{}

func StackList(rw web.ResponseWriter, req *web.Request) {

	// TODO: pluggable file storage
	var fileNames []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// this assumes a single directory
		if !info.IsDir() {
			fileNames = append(fileNames, strings.Replace(path, root, "", 1)[1:])
		}
		return nil
	})

	json.NewEncoder(rw).Encode(&fileNames)
}

func parseForm(rw web.ResponseWriter, req *web.Request) bool {
	if err := req.ParseForm(); err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return false
	}
	return true
}

func HeatMap(rw web.ResponseWriter, req *web.Request) {
	// /heatmap/?filename=perf.stacks01&rows=50
	if !parseForm(rw, req) {
		return
	}

	filename := req.FormValue("filename")
	absPath := filepath.Join(root, filename)

	rows, err := strconv.Atoi(req.FormValue("rows"))
	if err != nil {
		rows = 50
	}

	file, err := os.Open(absPath)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	offsets := heatmap.GenerateOffsets(bufio.NewScanner(file))
	hm := heatmap.GenerateHeatMap(offsets, rows)

	json.NewEncoder(rw).Encode(&hm)
}

func Stack(rw web.ResponseWriter, req *web.Request) {

	// /stack?filename=perf.stacks01&start=18.62&end=19.6
	if !parseForm(rw, req) {
		return
	}

	filename := req.FormValue("filename")
	absPath := filepath.Join(root, filename)
	file, err := os.Open(absPath)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stackStart, stackEnd := stack.CalculateStackRange(bufio.NewScanner(file))

	start, end := stackStart, stackEnd

	if req.FormValue("end") != "" {
		reqEnd := gfutil.MustParseFloat(req.FormValue("end"))
		if (stackStart + reqEnd) > stackEnd {
			fmt.Println("ERROR: ", stackStart, stackEnd, req.URL.String())
			rw.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		end = stackStart + reqEnd
	}
	if req.FormValue("start") != "" {
		reqStart := gfutil.MustParseFloat(req.FormValue("start"))
		start = start + reqStart
		if start > end {
			fmt.Println("ERROR: ", stackStart, stackEnd, req.URL.String())
			rw.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
	}

	// start and end are now the range that we want
	file, err = os.Open(absPath)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	root := stack.CreateFlameGraph(file, start, end)

	json.NewEncoder(rw).Encode(root)

}

func main() {
	rp := flag.String("root", "examples", "examples root")
	flag.Parse()

	if fi, err := os.Stat(*rp); err != nil {
		panic(err)
	} else {
		if !fi.IsDir() {
			panic("root is not a directory")
		}

		absPath, err := filepath.Abs(*rp)
		if err != nil {
			panic(err)
		}
		root = absPath
	}

	router := web.New(Context{})

	//option := web.StaticOption{IndexFile: "_index.html"}
	router.Middleware(web.LoggerMiddleware)
	sf, err := httpfs.NewStaticFileSystem("")
	if err != nil {
		panic(err)
	}
	mwf := web.StaticMiddlewareFromDir(sf)
	router.Middleware(func(w web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		// http.FileServer does a permanent redirect of /index.html to /
		// so for this path we serve /_index.html to avoid a redirect loop
		if req.URL.Path == "/" {
			req.URL.Path = "/_index.html"
		}
		mwf(w, req, next)
	})

	router.Get("/stack/list", StackList)
	router.Get("/heatmap/", HeatMap)
	router.Get("/stack", Stack)

	http.ListenAndServe(":8080", router)
}
