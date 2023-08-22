package es

import (
	"context"
	"errors"
	"strings"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"

	room2 "github.com/olaola-chat/rbp-proto/gen_pb/es/room"

	"github.com/olaola-chat/rbp-library/redis"
)

var esClient *Client

var EsSearch = &esSearch{}

type esSearch struct {
}

func init() {
	esClient = EsClientInit(EsVpc)
}

// BuildKeyword 关键词搜索
func BuildKeyword(keyword string) string {
	keyword = redis.EscapeTextFileString(strings.TrimSpace(keyword))
	return keyword
}

// QueryNoContent 不返回具体内容，只返回主键(出去前缀)
func (c *esSearch) QueryNoContent(ctx context.Context, index string, params map[string]interface{}, reply *room2.RepEsRoomSearchDefault) error {
	resp, err := esClient.Search(index, params)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("搜索失败")
	}
	reply.Total = uint32(resp.Hits.Total)
	for _, hit := range resp.Hits.Hits {
		reply.Data = append(reply.Data, gconv.Uint32(hit.ID))
	}
	return nil
}

// QueryWithContent 返回所有数据内容，需要自己对返回内容进行处理
func (c *esSearch) QueryWithContent(ctx context.Context, index string, params map[string]interface{}) (uint32, []map[string]interface{}, error) {
	resp, err := esClient.Search(index, params)
	if err != nil {
		return 0, nil, err
	}
	if resp == nil {
		return 0, nil, errors.New("搜索失败")
	}
	response := make([]map[string]interface{}, 0)
	for _, hit := range resp.Hits.Hits {
		response = append(response, hit.Source)
	}
	return uint32(resp.Hits.Total), response, nil
}

func SetAppId(appIds []uint32) map[string]interface{} {
	return map[string]interface{}{
		"terms": map[string]interface{}{
			"app_id": appIds,
		},
	}
}

func SetMatchName(keyword string) map[string]interface{} {
	return map[string]interface{}{
		"match": map[string]interface{}{
			"name": keyword,
		},
	}
}

func SetProperty(property string) map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			"property": property,
		},
	}
}

func BuildQuery(must []interface{}, limit uint32) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"size": limit,
	}
}

func (c *esSearch) Update(ctx context.Context, index string, primary uint64, params map[string]interface{}) error {
	return esClient.Update(index, primary, params)
}

func (c *esSearch) SetTerm(key string, value interface{}) g.Map {
	return g.Map{
		"term": g.Map{
			key: value,
		},
	}
}

func (c *esSearch) SetTerms(key string, value interface{}) g.Map {
	return g.Map{
		"terms": g.Map{
			key: value,
		},
	}
}

func (c *esSearch) SetRange(key string, v1, v2 int64) g.Map {
	gl := g.Map{}
	if v1 > 0 {
		gl["gte"] = v1
	}
	if v2 > 0 {
		gl["lte"] = v2
	}
	return g.Map{
		"range": g.Map{
			key: gl,
		},
	}
}

func (c *esSearch) BuildQuery(must []interface{}, offset, limit uint32, sort g.Array) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"size": limit,
	}
}

func (c *esSearch) SetSort(s string, s2 string) g.Map {
	return g.Map{
		s: g.Map{
			"order": s2,
		},
	}
}
