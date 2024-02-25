
# 介绍

### 存放日常项目中沉淀的公共库和工具，如系统xnotify的控制器库、http client请求的封装等

---

## 库功能如下：
```
｜_xnotify：监控文件的修改，对其允许和拒绝
｜_utils：存放好用的小工具或者函数
｜_httpclient: http client请求的封装，符合rest api资源的请求方式
```


---

## http client使用示例如下：
``` go
import github.com/HugoWw/common-library/httpclient

type testData1 struct {
    UserId int    `json:"userId"`
    Id     int    `json:"id"`
    Title  string `json:"title"`
    Bodys  string `json:"body"`
}

client, err := httpclient.NewHttpClient(http.DefaultClient, 30, "https://jsonplaceholder.typicode.com", 5)
if err != nil {
    fmt.Println("new http client error:", err)
    return
}

data1 := testData1{}
err = client.Get().Prefix("/v1").SetPath("/posts/1").Do(context.TODO()).Into(&data1)
if err != nil {
    // cerrors.ResponseForErrorReason(err)转换client请求错误信息是为常规error类型还是自定义错误类型
    fmt.Println("http client request err:", cerrors.ResponseForErrorReason(err))
    return
}

fmt.Printf("out put get data info:%+v\n", data1)
```

---

## xnotify使用示例如下：
``` go
import github.com/HugoWw/common-library/xnotify

var end chan bool = make(chan bool)
pInfoChan := make(chan xnotify.ProcInfo, 2)
fa, err := xnotify.NewFaNotify(end, pInfoChan)
if err != nil {
    fmt.Println("NewFaNotify error:", err)
    os.Exit(1)
}

// 只有在对文件注册的监控事件类型为"FAN_ACCESS_PERM|FAN_OPEN_PERM"类型的时候
// 需要把对这些事件类型允许,还是拒绝的结果写回到fanotify的文件描述符(即写回内核中)，从而判断进程是否有权限对文件的操作；
err = fa.AddMonitorFile("/tmp/dir1", xnotify.FAN_ACCESS|xnotify.FAN_CLOSE_WRITE|xnotify.FAN_OPEN|xnotify.FAN_MODIFY)
if err != nil {
    fmt.Println("NewFaNotify add monitor file error:", err)
}

go fa.MonitorFileEvents()
go func() {
    for v := range pInfoChan {
        fmt.Printf("Get Fanotify Event info about process info: %+v\n", v)
    }
}()

quit := make(chan os.Signal)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
select {
case <-end:
    fa.RemoveMonitor()
    fa.Close()
    close(pInfoChan)
case <-quit:
    fa.RemoveMonitor()
    fa.Close()
    close(pInfoChan)
}
```