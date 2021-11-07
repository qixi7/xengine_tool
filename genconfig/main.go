package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"text/template"
)

var fileWatch *regexp.Regexp // 模板文件文件正则
var dirWatch *regexp.Regexp  // 需要遍历的文件夹标记(有的文件夹不需要去遍历, 比如隐藏文件夹.git)
var templatePtr *template.Template

var (
	config string // 配置文件路径
	target string // 目标模板路径. 默认生成所有服务器的配置
)

func init() {
	flag.StringVar(&config, "config", "tool/genconfig/devconfig.json", "config file path")
	flag.StringVar(&target, "target", "all", "target file path")

	fileWatch, _ = regexp.Compile("(.*)\\.json\\.template$")
	dirWatch, _ = regexp.Compile("^[^\\.]")
}

func main() {
	flag.Parse()
	cfg := newConfigVar()
	if !cfg.LoadJsonFile(config) {
		return
	}

	GenConfig(".", cfg)
}

func GenConfig(pathname string, cfg configVar) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Printf("read dir fail, err=%v", err.Error())
		return
	}
	for _, fi := range rd {
		if fi.IsDir() && dirWatch.MatchString(fi.Name()) {
			fullDir := pathname + "/" + fi.Name()
			GenConfig(fullDir, cfg)
		} else if fromFile := pathname + "/" + fi.Name(); fileWatch.MatchString(fi.Name()) {
			genFileNam := pathname + "/" + fileWatch.FindStringSubmatch(fi.Name())[1] + ".json"
			genImpl(genFileNam, fromFile, cfg)
		}
	}
}

// 实际生成配置函数
func genImpl(genFileNam, fromFile string, cfg configVar) {
	var err error
	fmt.Printf("generate from file=%s, genFileName=%s\n", fromFile, genFileNam)
	templatePtr, err = template.ParseFiles(fromFile)
	if err != nil {
		fmt.Printf("Parse template fileName=%s, err=%v\n", fromFile, err.Error())
		return
	}
	// 先删除文件, 再创建文件
	_ = os.Remove(genFileNam)
	err = templatePtr.Execute(&fileWriter{
		file: genFileNam,
	}, cfg)
	if err != nil {
		fmt.Printf("Execute template fileName=%s, err=%v\n", fromFile, err.Error())
		return
	}
}

type fileWriter struct {
	file string
}

func (f *fileWriter) Write(p []byte) (int, error) {
	fileStream, err := os.OpenFile(f.file, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0600))
	if err != nil {
		return 0, err
	}
	fileStream.WriteString(string(p))
	fileStream.Close()
	return len(p), nil
}
