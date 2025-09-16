package es

import (
	"blog/internal/db"
	"context"
	"encoding/json"
	"strconv"

	elastic "github.com/olivere/elastic/v7"
)

type ES struct {
	client *elastic.Client
	idx    string
}

type ESDoc struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags,omitempty"`
}

func New(client *elastic.Client) *ES {
	e := &ES{client: client, idx: "posts"}
	// ensure index exists (simple)
	ctx := context.Background()
	exists, _ := client.IndexExists(e.idx).Do(ctx)
	if !exists {
		// create with default mapping (you can add analyzers)
		client.CreateIndex(e.idx).Do(ctx)
	}
	return e
}

func (e *ES) IndexPost(ctx context.Context, p *db.Post) error {
	doc := ESDoc{ID: p.ID, Title: p.Title, Content: p.Content, Tags: p.Tags}
	_, err := e.client.Index().
		Index(e.idx).
		Id(strconv.Itoa(p.ID)).
		BodyJson(doc).
		Do(ctx)
	return err
}

func (e *ES) Search(ctx context.Context, q string) ([]ESDoc, error) {
	matchQuery := elastic.NewMultiMatchQuery(q, "title", "content")
	res, err := e.client.Search().Index(e.idx).Query(matchQuery).Do(ctx)
	if err != nil {
		return nil, err
	}
	var out []ESDoc
	for _, hit := range res.Hits.Hits {
		var d ESDoc
		if err := json.Unmarshal(hit.Source, &d); err == nil {
			out = append(out, d)
		}
	}
	return out, nil
}

// Bonus: search related by tags
func (e *ES) RelatedByTags(ctx context.Context, tags []string, excludeID int, size int) ([]ESDoc, error) {
	boolQ := elastic.NewBoolQuery()
	if len(tags) > 0 {
		shoulds := make([]elastic.Query, 0, len(tags))
		for _, t := range tags {
			shoulds = append(shoulds, elastic.NewTermQuery("tags.keyword", t))
		}
		boolQ.Should(shoulds...)
		boolQ.MustNot(elastic.NewTermQuery("id", excludeID))
	}
	res, err := e.client.Search().Index(e.idx).Query(boolQ).Size(size).Do(ctx)
	if err != nil {
		return nil, err
	}
	var out []ESDoc
	for _, hit := range res.Hits.Hits {
		var d ESDoc
		if err := json.Unmarshal(hit.Source, &d); err == nil {
			out = append(out, d)
		}
	}
	return out, nil
}
