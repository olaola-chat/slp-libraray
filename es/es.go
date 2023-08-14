package es

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/frame/gins"
	"github.com/gogf/gf/util/gconv"
)

const (
	//HTTPGet 定义http get方法
	HTTPGet = "GET"
	//HTTPPost 定义http post方法
	HTTPPost = "POST"
	//HTTPDelete 定义http delete方法
	HTTPDelete = "DELETE"
)

const (
	//EsUser 用户,房间索引集群
	EsUser = "default"
	EsNew  = "es_new"
	EsRush = "rush"
	EsVpc  = "es_vpc"
)

// EsClient 单例创建一个Es client 对象
func EsClientInit(name string) *Client {
	instanceKey := fmt.Sprintf("self-go-es.%s", name)
	result := gins.GetOrSetFuncLock(instanceKey, func() interface{} {
		config := Config{}
		err := g.Cfg().GetStruct(fmt.Sprintf("go-es.%s", name), &config)
		fmt.Printf("xxxxx:%v", config)
		if err != nil {
			panic(gerror.Wrap(err, "es config error"))
		}
		return &Client{
			Config: &config,
		}
	})
	if client, ok := result.(*Client); ok {
		return client
	}
	//理论上是不可能到这一步的
	panic(gerror.New("get es client error"))
}

type EsInstance struct {
	EsUser *Client
	EsNew  *Client
	EsRush *Client
	EsVpc  *Client
}

var instance *EsInstance

// 初始化redis实例
func init() {
	g.Log().Info("init es instance")
	instance = &EsInstance{
		EsUser: EsClientInit(EsUser),
		EsNew:  EsClientInit(EsNew),
		EsRush: EsClientInit(EsRush),
		EsVpc:  EsClientInit(EsVpc),
	}
}

func EsClient(name string) *Client {
	switch name {
	case EsUser:
		return instance.EsUser
	case EsNew:
		return instance.EsNew
	case EsRush:
		return instance.EsRush
	case EsVpc:
		return instance.EsVpc

	default:
		panic(gerror.New("get es client error"))
	}
}

// Client 封装ES常用方法
type Client struct {
	Config *Config
}

// Int2String 数字文档ID转成字符文档ID
func (c *Client) Int2String(val interface{}) string {
	return gconv.String(val)
}

// IntSlice2StringSlice 数字文档ID转成字符文档ID
func (c *Client) IntSlice2StringSlice(val []interface{}) []string {
	return gconv.Strings(val)
}

// Search ES search
func (c *Client) Search(index string, body interface{}) (*SearchResponse, error) {
	res := &SearchResponse{}
	url := fmt.Sprintf("%s/_search", index)
	err := c.exec(url, HTTPPost, body, res)
	return res, err
}

// Mget ES 批量获取文档
func (c *Client) Mget(index string, ids []string, fields ...[]string) (*MResponse, error) {
	res := &MResponse{}
	if len(ids) == 0 {
		return res, nil
	}
	url := fmt.Sprintf("%s/default/_mget", index)
	if len(fields) > 0 && len(fields[0]) > 0 {
		url = fmt.Sprintf("%s?_source=%s", url, strings.Join(fields[0], ","))
	}
	err := c.exec(url, HTTPPost, ids, res)
	return res, err
}

// Get ES 获取单个文档
func (c *Client) Get(index, docID string, fields ...[]string) (*GetResponse, error) {
	res := &GetResponse{}
	url := fmt.Sprintf("%s/default/%s", index, docID)
	if len(fields) > 0 && len(fields[0]) > 0 {
		url = fmt.Sprintf("%s?_source=%s", url, strings.Join(fields[0], ","))
	}
	err := c.execGet(url, res)
	return res, err
}

func (c *Client) Put(index, docID string, data map[string]interface{}) error {
	res := &PutResponse{}
	url := fmt.Sprintf("%s/default/%s", index, docID)
	err := c.exec(url, HTTPPost, data, res)
	return err
}

func (c *Client) Update(index string, id uint64, data map[string]interface{}) error {
	res := &PutResponse{}
	url := fmt.Sprintf("%s/default/%d/_update", index, id)
	updateData := make(map[string]interface{})
	updateData["doc"] = data
	err := c.exec(url, HTTPPost, updateData, res)
	return err
}

func (c *Client) Delete(index string, id uint64) error {
	res := &PutResponse{}
	url := fmt.Sprintf("%s/default/%d", index, id)
	err := c.exec(url, HTTPDelete, nil, res)
	return err
}

func (c *Client) execGet(url string, pointer interface{}) error {
	return c.exec(url, HTTPGet, nil, pointer)
}

func (c *Client) exec(url string, method string, data interface{}, pointer interface{}) error {
	var body []byte
	queryURL := fmt.Sprintf("http://%s:%d/%s", c.Config.Host, c.Config.Port, url)
	datastr, _ := json.Marshal(data)

	client := g.Client()
	//client.Use(model.TraceHTTPClient)
	client.SetTimeout(time.Second)
	client.SetBasicAuth(c.Config.User, c.Config.Password)
	if method == HTTPPost {
		xClient := client.ContentType("application/json")
		if data == nil {
			body = xClient.PostBytes(queryURL)
		} else {
			body = xClient.PostBytes(queryURL, data)
		}
	} else if method == HTTPGet {
		body = client.GetBytes(queryURL)
	} else {
		return gerror.New("method error")
	}

	if len(body) == 0 {
		return gerror.New("http request error")
	}
	g.Log().Printf("exec:%v, %v, %v", queryURL, string(datastr), string(body))
	normalError := json.Unmarshal(body, pointer)
	if normalError == nil {
		return nil
	}

	//尝试解析是不是ErrorResponse
	errorRes := &ErrorResponse{}
	err := json.Unmarshal(body, errorRes)
	if err == nil {
		if errorRes.Error.Reason == "" {
			return normalError
		}
		return gerror.New(errorRes.Error.Reason)
	}

	return normalError
}
