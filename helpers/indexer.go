package helpers

import (
	"AinedIndexItemCLI/databases"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

// Indexer оборачивает клиента Elasticsearch для работы с индексами
type Indexer struct {
	Client *databases.Client
}

// NewIndexer создает новый экземпляр Indexer
func NewIndexer(client *databases.Client) *Indexer {
	return &Indexer{Client: client}
}

// IndexData индексирует данные в указанном индексе
func (indexer *Indexer) IndexData(index string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data to JSON: %v", err)
	}
	if documentID, err := uuid.NewV7(); err == nil {
		req := esapi.IndexRequest{
			Index:      index,
			DocumentID: documentID.String(),
			Body:       bytes.NewReader(jsonData),
		}

		res, err := req.Do(context.Background(), indexer.Client.ES)
		if err != nil {
			return fmt.Errorf("error indexing document: %v", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("error response from Elasticsearch: %s", res.String())
		}

		//log.Printf("Document %s indexed successfully to index %s", documentID, index)
	}

	return nil
}

// DeleteOldDocuments удаляет документы, которые соответствуют заданному условию
func (indexer *Indexer) DeleteOldDocuments(index string) error {

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{}, // Запрос для соответствия всем документам
		},
	}

	// Преобразование запроса в JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error marshalling query to JSON: %v", err)
	}

	// Создание запроса на удаление
	req := esapi.DeleteByQueryRequest{
		Index: []string{index},
		Body:  bytes.NewReader(queryJSON),
	}

	// Выполнение запроса на удаление
	res, err := req.Do(context.Background(), indexer.Client.ES)
	if err != nil {
		return fmt.Errorf("error deleting documents by query: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	//res, err := indexer.Client.ES.Search(
	//	indexer.Client.ES.Search.WithContext(context.Background()),
	//	indexer.Client.ES.Search.WithIndex(index),
	//	indexer.Client.ES.Search.WithPretty(),
	//
	//)
	//
	//if err != nil {
	//	return fmt.Errorf("error getting response from Elasticsearch: %v", err)
	//}
	//defer res.Body.Close()
	//
	//if res.IsError() {
	//	return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	//}
	//
	//var result map[string]interface{}
	//if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
	//	return fmt.Errorf("error parsing the response body: %v", err)
	//}
	//
	//hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	//fmt.Println(hits)
	//chunks := InterfChunkSlice(hits, 1000)
	//var wg sync.WaitGroup
	//var mu sync.Mutex
	//for _, chunk := range chunks {
	//	wg.Add(1)
	//	go func(chunk []interface{}) {
	//		defer wg.Done()
	//		for _, hit := range chunk {
	//
	//			docID := hit.(map[string]interface{})["_id"].(string)
	//			req := esapi.DeleteRequest{
	//				Index:      index,
	//				DocumentID: docID,
	//			}
	//
	//			delRes, err := req.Do(context.Background(), indexer.Client.ES)
	//
	//			if err != nil {
	//				mu.Lock()
	//				log.Printf("error deleting document %s: %v", docID, err)
	//				mu.Unlock()
	//				return
	//			}
	//			defer delRes.Body.Close()
	//
	//			if delRes.IsError() {
	//				mu.Lock()
	//				log.Printf("error response from Elasticsearch while deleting document %s: %s", docID, delRes.String())
	//				mu.Unlock()
	//			} else {
	//				mu.Lock()
	//				log.Printf("Document %s deleted successfully from index %s", docID, index)
	//				mu.Unlock()
	//			}
	//
	//		}
	//	}(chunk)
	//}
	//wg.Wait()

	//for _, hit := range hits {
	//	docID := hit.(map[string]interface{})["_id"].(string)
	//	req := esapi.DeleteRequest{
	//		Index:      index,
	//		DocumentID: docID,
	//	}
	//
	//	delRes, err := req.Do(context.Background(), indexer.Client.ES)
	//	if err != nil {
	//		log.Printf("error deleting document %s: %v", docID, err)
	//		continue
	//	}
	//	defer delRes.Body.Close()
	//
	//	if delRes.IsError() {
	//		log.Printf("error response from Elasticsearch while deleting document %s: %s", docID, delRes.String())
	//	} else {
	//		log.Printf("Document %s deleted successfully from index %s", docID, index)
	//	}
	//}

	return nil
}

// SearchData выполняет поиск по индексу
func (indexer *Indexer) SearchData(index string, query map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %v", err)
	}

	res, err := indexer.Client.ES.Search(
		indexer.Client.ES.Search.WithContext(context.Background()),
		indexer.Client.ES.Search.WithIndex(index),
		indexer.Client.ES.Search.WithBody(&buf),
		indexer.Client.ES.Search.WithTrackTotalHits(true),
		indexer.Client.ES.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting response from Elasticsearch: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %v", err)
	}

	return result, nil
}
