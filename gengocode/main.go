package main

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"text/template"
)

type config struct {
	SrcDir   string
	VisitDir []string
}

// config
var visitConfig = &config{}

var fileWatch *regexp.Regexp // 模板文件文件正则
var dirWatch *regexp.Regexp  // 需要遍历的文件夹标记(有的文件夹不需要去遍历, 比如隐藏文件夹.git)
var templatePtr *template.Template

const _structCreatorFuncName = "isTypeCreator" // struct creator
const _pbCreatorFuncName = "ProtoMessage"      // protobuf creator. 这里偷下懒,,,

func init() {
	var err error
	fileWatch = regexp.MustCompile(`(.*)\.go$`)
	dirWatch = regexp.MustCompile(`^[^\.]`)
	templatePtr, err = template.New("creator").Parse(creatorTemplate)
	if err != nil {
		log.Fatalf("Parse template  err=%v", err)
	}
	// load
	loadFunc := func(path string, cfg *config) bool {
		f, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("load json file, path=%s, err=%v\n", path, err)
			return false
		}
		err = jsoniter.Unmarshal(f, cfg)
		if err != nil {
			fmt.Printf("Unmarshal json file, path=%s, err=%v\n", path, err)
			return false
		}
		return true
	}
	if !loadFunc("./config.json", visitConfig) {
		os.Exit(2)
	}
}

// Source source
type source struct {
	Fset *token.FileSet
	F    *ast.File
}

// NewSource new source
func newSource(fileName string) *source {
	s := &source{
		Fset: token.NewFileSet(),
	}
	f, err := parser.ParseFile(s.Fset, fileName, nil, 0)
	if err != nil {
		log.Fatalf("无法解析源文件: %s", fileName)
	}
	s.F = f
	return s
}

func parse(s *source, bodyBuffer, packageBuff *bytes.Buffer, structSlice *[]string) {
	for _, oneDecl := range s.F.Decls {
		switch decl := oneDecl.(type) {
		case *ast.FuncDecl:
			// 函数定义
			if decl.Recv == nil || (decl.Name.Name != _structCreatorFuncName && decl.Name.Name != _pbCreatorFuncName) {
				// 非接收者函数||定义不是creator标记 ===> 直接滚蛋~
				continue
			}

			if length := len(decl.Recv.List); length > 1 {
				log.Printf("len(decl.Recv.List)=%d", length)
			}

			var receiver string
			switch recvType := decl.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				receiver = recvType.X.(*ast.Ident).Name
			case *ast.Ident:
				receiver = recvType.Name
			default:
				log.Fatalf("UnHandled recvType=%v", recvType)
			}
			if objTemp := s.F.Scope.Lookup(receiver); objTemp == nil {
				log.Printf("receiver=%s not in objMap", receiver)
				continue
			}
			if packageBuff.Len() == 0 {
				packageBuff.WriteString(s.F.Name.Name)
			}
			*structSlice = append(*structSlice, receiver)
			generateCode(receiver, bodyBuffer)
		}
	}
}

func walkCode(pathname string) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		log.Printf("read dir fail, err=%v", err.Error())
		return
	}
	bodyBuffer := &bytes.Buffer{}
	packageBuff := &bytes.Buffer{}
	structBuff := make([]string, 0, 5)
	for _, fi := range rd {
		if fi.IsDir() && dirWatch.MatchString(fi.Name()) {
			// 如果是合法文件夹. 继续向下遍历
			walkCode(pathname + "/" + fi.Name())
		} else if fullName := pathname + "/" + fi.Name(); fileWatch.MatchString(fi.Name()) {
			parse(newSource(fullName), bodyBuffer, packageBuff, &structBuff)
		}
	}
	if packageBuff.Len() == 0 {
		return
	}
	genFileName := fmt.Sprintf("%s/gen_%s.go", pathname, packageBuff.String())
	log.Printf("genFileName=%s", genFileName)
	codeBuff := bytes.Buffer{}
	// 加上header
	codeBuff.WriteString(headerTemplate)
	// 构建package
	codeBuff.WriteString(fmt.Sprintf("package %s\n\n", packageBuff.String()))
	// 构建struct
	codeBuff.WriteString("type smallObjCreatorType struct {\n")
	for _, structInfo := range structBuff {
		codeBuff.WriteString(fmt.Sprintf("\tcreator_%s\n", structInfo))
		codeBuff.WriteString(fmt.Sprintf("\tcreatorSlice_%s\n", structInfo))
		codeBuff.WriteString(fmt.Sprintf("\tcreatorPtr_%s\n", structInfo))
	}
	codeBuff.WriteString("}\n\n")
	codeBuff.WriteString("var smallObjCreator smallObjCreatorType\n")
	// 构建code body
	codeBuff.WriteString(bodyBuffer.String())
	if err = ioutil.WriteFile(genFileName, codeBuff.Bytes(), 0644); err != nil {
		log.Fatalf("writeFile %s, err=%v", genFileName, err)
	}
}

func generateCode(structName string, bodyBuffer *bytes.Buffer) {
	err := templatePtr.Execute(bodyBuffer, &codeConfig{
		Name:  structName,
		TName: structName,
	})
	if err != nil {
		log.Fatalf("Execute template, err=%v", err)
	}
}

type codeConfig struct {
	Name  string
	TName string
}

func main() {
	for _, targetDir := range visitConfig.VisitDir {
		walkCode(visitConfig.SrcDir + targetDir)
	}
}
