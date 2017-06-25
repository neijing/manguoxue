package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"html/template"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	templateList = "list.html"
	templatePage = "page.html"
)

var (
	root         = flag.String("root", "/home/lexc/go/3rd/manguoxue", "data directory")
	sdir         = flag.String("D", ".", "data directory")
	bMp3         = flag.Bool("mp3", false, "whether using baidu to generate mp3 or not?")
	allInOne     = flag.Bool("one", false, "whether all in one page?")
	listMap      map[string][]string
	detailMap    map[string]string
	tomlMap      map[string]tomlFile
	contentPath  string
	publicPath   string
	templatePath string
)

// {"shortUrl":"http://dwz.cn/61UT1M","status":"success"}
type BaiduVoice struct {
	Url    string `json:"shortUrl"`
	Status string `json:"status"`
}

type tomlFile struct {
	Name string `toml:"name"`
	Desc string `toml:"desc"`
}

type listHtml struct {
	Navi    template.HTML
	Content template.HTML
}

type pageHtml struct {
	Navi         template.HTML
	SectionTitle template.HTML
	Content      template.HTML
	Mp3Url       template.HTML
	PrePage      template.HTML
	NextPage     template.HTML
}

func init() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	listMap = make(map[string][]string)
	tomlMap = make(map[string]tomlFile)
	detailMap = make(map[string]string)
	contentPath = filepath.Join(*root, "content")
	publicPath = filepath.Join(*root, "www")
	templatePath = filepath.Join(*root, "template")
	return
}

func getBaiduMp3Url(sectiontitle string, content string) (string, error) {
	cmd := fmt.Sprintf(`curl 'http://developer.baidu.com/vcast/getVcastInfo' -H 'Cookie: FP_UID=1f88fc3fd51aabedac21dc7b34416a37; BDUSS=jlVeExFamQ2R3FlaS1UWXpVQjhzY25FTGcxVW1MZmFrUHBzbzdORkJMSlI3MkpaSVFBQUFBJCQAAAAAAAAAAAEAAAD0Pei3b25la25pZmVlAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFFiO1lRYjtZf; BAIDUID=0D4FA5BCFBE334AB0CAF20A35B9FC26C:FG=1; PSTM=1497144976; H_PS_PSSID=1449_21094_21931; BIDUPSID=4F64597E899E867CD8C6441CC5BFAF5B; BDORZ=FFFB88E999055A3F8A630C64834BD6D0; pgv_pvi=8231539712; pgv_si=s7966715904; Hm_lvt_3abe3fb0969d25e335f1fe7559defcc6=1496418443,1497020650,1497145682; Hm_lpvt_3abe3fb0969d25e335f1fe7559defcc6=1497145682' -H 'Origin: http://developer.baidu.com' -H 'Accept-Encoding: gzip, deflate' -H 'Accept-Language: zh-CN,zh;q=0.8' -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36' -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' -H 'Accept: application/json, text/javascript, */*; q=0.01' -H 'Referer: http://developer.baidu.com/vcast' -H 'X-Requested-With: XMLHttpRequest' -H 'Connection: keep-alive' --data 'title=%s&content=%s&sex=0&speed=3&volumn=7&pit=5&method=TRADIONAL' --compressed`, url.PathEscape(sectiontitle), url.PathEscape(content))
	// log.Printf("cmd: %s", cmd)
	out, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		log.Print("error: ", err.Error())
		return "", err
	}
	log.Printf("baidu out: %s", string(out))
	var bv BaiduVoice
	if err := json.Unmarshal(out, &bv); err != nil {
		log.Print("error: ", err.Error())
		return "", err
	}
	return bv.Url, nil
}

func prepareNavi(htmlPath string) string {
	navi := ""
	s := ""
	hp := strings.Split(htmlPath, "/")
	front := "/"
	for _, pt := range hp {
		// fmt.Sprintf(`<li><a href="%s">%s</a><span class="divider">/</span></li>`
		log.Print(pt)
		if pt == "" {
			s = fmt.Sprintf(`<li><a href="%s">%s</a></li>`, front, "全部")
		} else {
			front = filepath.Join(front, pt)
			if v1, ok := tomlMap[front]; ok {
				s = fmt.Sprintf(`<li><a href="%s">%s</a></li>`, front, v1.Name)
			} else {
				log.Fatal(front, " can't be found in map")
			}
		}
		navi += s
	}
	return navi
}

func contentPage(path string, htmlPath string) error {
	cmd := fmt.Sprintf("/usr/bin/pandoc -f markdown -t html %s", path)
	log.Printf("cmd: %s", cmd)
	out, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		log.Print("error: ", err.Error())
		return err
	}
	t, err := template.ParseFiles(filepath.Join(templatePath, templatePage))
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return err
	}
	navi := prepareNavi(htmlPath)
	// write html file
	nf := strings.Split(filepath.Base(path), ".")
	fileName := fmt.Sprintf("%s.html", nf[0])
	nf0, err := strconv.Atoi(nf[0])
	preNo := nf0 - 1
	var prepage, nextpage string
	if preNo == 0 {
		prepage = "#"
	} else {
		prepage = fmt.Sprintf("%s/%02d.html", htmlPath, preNo)
	}
	nextNo := nf0 + 1
	nextpage = fmt.Sprintf("%s/%02d.html", htmlPath, nextNo)
	// record href
	detailPath := filepath.Join(htmlPath, fileName)
	ib := strings.Index(string(out), ">")
	ie := strings.Index(string(out), "</h1>")
	sectionTitle := string(out)[ib+1 : ie]
	detailMap[detailPath] = sectionTitle

	file, err := os.Open(path) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	le := strings.Index(string(b), "\n")
	if le < 0 {
		le = 0
	}
	var mp3url string
	if *bMp3 {
		mp3url, err = getBaiduMp3Url(sectionTitle, string(b[le:]))
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			return err
		}
	}
	if *allInOne {
		prepage = "#"
		nextpage = "#"
	}
	// public it
	pageData := pageHtml{
		Navi:         template.HTML(navi),
		SectionTitle: template.HTML(sectionTitle),
		Mp3Url:       template.HTML(mp3url),
		PrePage:      template.HTML(prepage),
		NextPage:     template.HTML(nextpage),
		Content:      template.HTML(string(out)),
	}

	pubPath := filepath.Join(publicPath, htmlPath, fileName)
	// write html file
	if err := os.MkdirAll(filepath.Dir(pubPath), 0755); err != nil {
		log.Print("error: ", err.Error())
		return err
	}
	f, err := os.OpenFile(pubPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(f, pageData)
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	// prepare index.html
	v, ok := listMap[htmlPath]
	if ok {
	} else {
		v = make([]string, 0, 128)
	}
	v = append(v, fileName)
	listMap[htmlPath] = v
	return nil
}

func listPage(path string, htmlPath string) error {
	if htmlPath == "/" {
		return nil
	}
	dir := filepath.Dir(htmlPath)
	v, ok := listMap[dir]
	if ok {
	} else {
		v = make([]string, 0, 128)
	}
	v = append(v, filepath.Base(path))
	listMap[dir] = v
	log.Printf("htmlPath=%s, dir=%s, v=%v", htmlPath, dir, v)
	return nil
}

func walkFnDir(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	var htmlPath string
	switch filepath.Ext(path) {
	case ".toml":
		var of tomlFile
		if _, err := toml.DecodeFile(path, &of); err != nil {
			log.Print("error: ", err.Error())
			return err
		}
		htmlPath = path[len(contentPath):]
		htmlPath = filepath.Dir(htmlPath)
		tomlMap[htmlPath] = of
		log.Printf("htmlPath:%s, of:%v", htmlPath, of)
	case "":
		// directory
		if path == contentPath {
			htmlPath = "/"
		} else {
			htmlPath = path[len(contentPath):]
		}
		return listPage(path, htmlPath)
	}
	return err
}

func walkFnFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	htmlPath := path[len(contentPath):]
	if htmlPath == "" {
		htmlPath = "/"
	}
	htmlPath = filepath.Dir(htmlPath)
	switch filepath.Ext(path) {
	case ".md":
		return contentPage(path, htmlPath)
	}
	return err
}

func genFunc(sourcedir string) {
	if err := filepath.Walk(contentPath, walkFnDir); err != nil {
		log.Print("error: ", err.Error())
		return
	}
	if err := filepath.Walk(sourcedir, walkFnFile); err != nil {
		log.Print("error: ", err.Error())
		return
	}
	// generate all index.html
	t, err := template.ParseFiles(filepath.Join(templatePath, templateList))
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	for k, v := range listMap {
		navi := prepareNavi(k)
		var content string
		for _, directory := range v {
			href := filepath.Join(k, directory)
			if v2, ok := tomlMap[href]; ok {
				oneFile := fmt.Sprintf(`<li style="line-height:2"><a href="%s">%s</a></li>`, href, v2.Name)
				content += oneFile
			} else {
				// its a detail html page
				if v3, ok := detailMap[href]; ok {
					oneFile := fmt.Sprintf(`<li style="line-height:2"><a href="%s">%s</a></li>`, href, v3)
					content += oneFile
				} else {
					log.Fatal("href=%s, can't be found in map", href)
				}
			}
		}
		lh := listHtml{
			Navi:    template.HTML(navi),
			Content: template.HTML(content),
		}
		// generate index.html
		pubPath := filepath.Join(publicPath, k, "index.html")
		// write html file
		if err := os.MkdirAll(filepath.Dir(pubPath), 0755); err != nil {
			log.Print("error: ", err.Error())
			return
		}
		f, err := os.OpenFile(pubPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = t.Execute(f, lh)
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}
	return
}

func main() {
	genFunc(filepath.Join(contentPath, *sdir))
	return
}
