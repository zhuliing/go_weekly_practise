package main

import (
	"net/http"
	"os/signal"
	"os"
	"context"
	"time"
	"log"
	"fmt"
	"math/rand"
	"encoding/json"
)

/*
1.1 作业如下：

在第一周接收请求作业基础上，要求如下：
将接收到请求相关的业务日志异步方式写入文件
接收系统关闭指令，并可以做到平滑关闭
1.2作业目标：

协程应用
思考如何平滑重启
 */
var quit = make(chan os.Signal, 1)
var requestStatusMap = map[int]bool{}
var done  = make(chan os.Signal, 1)
var chanLog = make(chan map[string]interface{})

func main() {
	 http.HandleFunc("/hello", helloWorld)
	 signal.Notify(quit, os.Interrupt)
	 server := &http.Server{
	 	Addr:"127.0.0.1:8777",
	 	Handler:nil,
	 }
	 go killProcess(server, quit)
	 server.ListenAndServe()
	 <-done
}

func killProcess(server *http.Server, quit <- chan os.Signal)  {
	<-quit
	go shutDown(server)
	for  {
		if len(requestStatusMap) != 0 {
			fmt.Print("目前还有进行中的请求，请稍等\n")
			time.Sleep(time.Second * 2)
			continue
		} else {
			close(done)
			break
		}
	}
}
 
func shutDown(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); nil != err {
		log.Fatalf("server shutdown failed, err: %v\n", err)
	}
	log.Println("server gracefully shutdown")
}

func generateRangeNum(min int, max int) int {
	if min == max {
		return min
	}
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	go writeLog()
	var uniqueId = generateRangeNum(1,1000)
	requestStatusMap[uniqueId] = false
	reqMethod := r.Method
	reqUrl := r.URL.Path
	reqQuery := r.URL.RawQuery
	params := map[string]interface{} {
		"method" : reqMethod,
		"url" : reqUrl,
		"query" : reqQuery,
		"response" : "hello kitty, hello world~~~xixiixixix",
	}
	chanLog<-params
	w.Write([]byte("kitty ,kitty"))
}

func writeLog() {
	content := <-chanLog
	contentToJson,_ := json.Marshal(content)
	fileName := "/tmp/kitty.log"
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