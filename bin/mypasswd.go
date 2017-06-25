package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
)

var (
	addr = flag.String("bind", ":6666", "the http server address")
)

func init() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	return
}

func genPasswdFunc(w http.ResponseWriter, req *http.Request) {
	token := req.FormValue("t")
	if "" == token {
		log.Print(req)
		http.Error(w, "need t parameter", 403)
		return
	}
	site := req.FormValue("s")
	if "" == site {
		log.Print(req)
		http.Error(w, "need s parameter", 403)
		return
	}
	cryp := req.FormValue("c")
	if "" == cryp {
		cryp = "md5"
	}
	key := fmt.Sprintf("%s@%s", token, site)
	passwd := ""
	switch cryp {
	case "md5":
		passwd = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	case "sha1":
		passwd = fmt.Sprintf("%x", sha1.Sum([]byte(key)))
	case "sha256":
		passwd = fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	case "sha512":
		passwd = fmt.Sprintf("%x", sha512.Sum512([]byte(key)))
	default:
		passwd = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	}
	io.WriteString(w, passwd[:13])
	return
}

func main() {
	http.HandleFunc("/gp", genPasswdFunc)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("mypasswd failed, error: ", err)
	}
}
