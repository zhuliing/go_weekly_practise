package question1

/**
在第一周接收请求作业基础上，要求如下：
将接收到请求相关的业务日志异步方式写入文件
接收系统关闭指令，并可以做到平滑关闭
 */

import (
	"net/http"
	"encoding/json"
	"os"
)

var chanLog = make(chan map[string]interface{})
func main() {
	http.HandleFunc("/hello", helloWorld)
	http.ListenAndServe("127.0.0.1:8777", nil)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	go writeLog()
	reqMethod := r.Method
	reqUrl := r.URL.Path
	reqQuery := r.URL.RawQuery
	params := map[string]interface{} {
		"method" : reqMethod,
		"url" : reqUrl,
		"query" : reqQuery,
		"response" : "hello kitty, hello world~~~",
	}
	chanLog<-params
	w.Write([]byte("ahhahahaha"))
}

func writeLog() {
	content := <-chanLog
	contentToJson,_ := json.Marshal(content)
	fileName := "/tmp/hello.log"
	_,err := os.Stat(fileName)
	if err != nil || os.IsNotExist(err){
		//创建一个文件，文件mode是0666(读写权限),如果文件已经存在，则重新创建一个，原文件被覆盖，创建的新文件具有读写权限,
		// 等同于OpenFile(name string, O_CREATE,0666)
		os.Create(fileName)
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0)
	defer file.Close()
	//seek设置下一次读或写操作的偏移量offset，
	// 根据whence来解析：0意味着相对于文件的原始位置，1意味着相对于当前偏移量，2意味着相对于文件结尾。它返回新的偏移量和错误（如果存在）
	index,_ := file.Seek(0,2)
	file.WriteAt([]byte(string(contentToJson)  + "\n"), index)
}

