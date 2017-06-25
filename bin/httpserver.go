package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	templateList = "list.html"
)

type listHtml struct {
	Navi    template.HTML
	Content template.HTML
}

var (
	addr         = flag.String("bind", ":8066", "the http server address")
	root         = flag.String("root", "/home/lexc/go/3rd/wnq", "data directory")
	templatePath string
	publicPath   string
)

func init() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	publicPath = filepath.Join(*root, "www")
	templatePath = filepath.Join(*root, "template")
	return
}

func SearchFunc(w http.ResponseWriter, req *http.Request) {
	key := req.FormValue("k")
	if "" == key {
		log.Print(req)
		http.Error(w, "need k parameter", 403)
		return
	}
	cmd := fmt.Sprintf(`grep -R "%s" %s`, key, publicPath)
	// log.Printf("cmd: %s", cmd)
	out, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		log.Print("error: ", err.Error())
		http.Error(w, "internal error", 500)
		return
	}
	log.Printf("baidu out: %s", string(out))
	t, err := template.ParseFiles(filepath.Join(templatePath, templateList))
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	// generate search result
	idx := 1
	content := ""
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, ":<p>") {
			ts := strings.Split(txt, ":")
			oneFile := fmt.Sprintf(`<li style="line-height:2"><a href="%s">>>%d</a>%s</li>`,
				strings.Replace(ts[0], publicPath, "", 1), idx, ts[1])
			content += oneFile
			idx++
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print("error: ", err)
		return
	}
	lh := listHtml{
		Navi:    template.HTML(key),
		Content: template.HTML(content),
	}
	err = t.Execute(w, lh)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func main() {
	http.HandleFunc("/search.cgi", SearchFunc)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
