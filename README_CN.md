# httprpc

这是一个http请求的包，有一个Context对象是对一个请求从生成到响应以及处理响应的描述。所有的请求和响应都会以Context为基础来实现。

函数申明如下所示:
```golang
type Context struct{
    *http.Response  //Do执行成功后才会被填充
}
//以下函数均描述了*Context的生成过程
//Get和Post都是在net/http的方法前加了一个context.Context参数
//JSON和XML都是Post自动加上Context-Type的语法糖函数
func Request(ctx context.Context, req *http.Request) (c *Context)
func Get(ctx context.Context, url string) (c *Context)
func Post(ctx context.Context, url, contentType string, body io.Reader) (c *Context)
func JSON(ctx context.Context, url string, i interface{}) *Context
func XML(ctx context.Context, url string, i interface{}) *Context

//Do会执行中间件并请求数据,所以在Do()之前`http.Response`都为nil
//Do返回自己的`Context`方便执行后面的函数调用
func (c *Context) Do() *Context {}

//以下函数都是快速获取不同格式的返回值的语法糖
//Bytes后面的函数都会调用Bytes函数获取http.Response的返回,所以以下函数都不用考虑io.Reader读取问题
//如果请求的是一个文件,调用下面函数的时候可能要考虑内存问题，可以在Do()之后直接对*http.Response进行读取
func (c *Context) Bytes() (bs []byte, err error) {}
func (c *Context) String() (str string, err error) {}
func (c *Context) IntoJSON(v interface{}) (err error) {}
func (c *Context) IntoXML(v interface{}) (err error) {}
```

## 用法

#### 简单的例子

```golang
ctx := context.TODO()
body, err := httprpc.Get(ctx, "https://enrprehryqtc.x.pipedream.net").String()
if err != nil {
    t.Error(err)
}
t.Logf("resp: %s", string(body))
```

#### 请求json的例子

```golang
ctx := context.TODO()
var ret = make(map[string]interface{})
err := httprpc.JSON(ctx, "https://webgfw2.ymt.com/pub/v10/appim/default/websocket_reg.json?app_key=4001&fCode=1000002", nil).IntoJSON(&ret)
if err != nil {
    t.Error(err)
}
t.Logf("resp: %v", ret)
```


#### 获取响应状态

```golang
ctx := context.TODO()
if httpCtx,err := httprpc.JSON(ctx, "https://baidu.com", nil).Do(); err != nil {
	t.Error(httpCtx.Err())
}
status:=httpCtx.StatusCode
```

#### 高级用法
Context是对一个请求从生成到响应以及处理响应的描述。在这中间一定有一些公共的处理方法，比如添加日志，设置超时时间，添加代理等。这些需求都可以通过一个叫中间件的方式来做。如有需要，请看源码。另外可以参考[accesslog中间件](access.go)

