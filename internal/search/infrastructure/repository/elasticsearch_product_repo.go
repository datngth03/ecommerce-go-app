// internal/search/infrastructure/repository/elasticsearch_product_repo.go
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/datngth03/ecommerce-go-app/internal/search/domain" // Import domain package
)

const (
	indexName = "products" // Tên chỉ mục trong Elasticsearch
)

// ElasticsearchProductRepository implements the domain.SearchProductRepository interface.
// ElasticsearchProductRepository triển khai giao diện domain.SearchProductRepository.
type ElasticsearchProductRepository struct {
	es *elasticsearch.Client
}

// NewElasticsearchProductRepository creates a new instance of ElasticsearchProductRepository.
// NewElasticsearchProductRepository tạo một thể hiện mới của ElasticsearchProductRepository.
func NewElasticsearchProductRepository(esClient *elasticsearch.Client) *ElasticsearchProductRepository {
	repo := &ElasticsearchProductRepository{es: esClient}
	// Khởi tạo chỉ mục (index) nếu chưa tồn tại
	repo.createProductIndex()
	return repo
}

// createProductIndex checks if the 'products' index exists and creates it if not.
// createProductIndex kiểm tra xem chỉ mục 'products' có tồn tại không và tạo nó nếu chưa.
func (r *ElasticsearchProductRepository) createProductIndex() {
	res, err := r.es.Indices.Exists([]string{indexName})
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		log.Printf("Elasticsearch index '%s' already exists.", indexName)
		return
	}

	// Define the mapping for the 'products' index
	// Định nghĩa mapping cho chỉ mục 'products'
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		},
		"mappings": {
			"properties": {
				"id":            {"type": "keyword"},
				"name":          {"type": "text", "analyzer": "standard", "fields": {"keyword": {"type": "keyword"}}},
				"description":   {"type": "text", "analyzer": "standard"},
				"price":         {"type": "double"},
				"category_id":   {"type": "keyword"},
				"image_urls":    {"type": "keyword"},
				"stock_quantity":{"type": "integer"},
				"created_at":    {"type": "date"},
				"updated_at":    {"type": "date"}
			}
		}
	}`

	res, err = r.es.Indices.Create(indexName, r.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		log.Fatalf("Error creating index '%s': %s", indexName, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error response creating index '%s': %s", indexName, res.String())
	}

	log.Printf("Elasticsearch index '%s' created successfully.", indexName)
}

// IndexProduct adds or updates a product in the search engine.
// IndexProduct thêm hoặc cập nhật một sản phẩm vào công cụ tìm kiếm.
func (r *ElasticsearchProductRepository) IndexProduct(ctx context.Context, product *domain.SearchProduct) error {
	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal product: %v", domain.ErrFailedToIndexProduct, err)
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: product.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true", // Refresh the index immediately for testing
	}

	res, err := req.Do(ctx, r.es)
	if err != nil {
		return fmt.Errorf("%w: failed to perform index request: %v", domain.ErrFailedToIndexProduct, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("%w: failed to parse ES error response: %v", domain.ErrFailedToIndexProduct, err)
		}
		return fmt.Errorf("%w: Elasticsearch index error: [%s] %s", domain.ErrFailedToIndexProduct, res.Status(), e["error"])
	}

	log.Printf("Indexed product %s", product.ID)
	return nil
}

// SearchProducts searches for products based on a query and filters.
// SearchProducts tìm kiếm sản phẩm dựa trên truy vấn và bộ lọc.
func (r *ElasticsearchProductRepository) SearchProducts(ctx context.Context, query, categoryID string, minPrice, maxPrice float64, limit, offset int32) ([]*domain.SearchProduct, int64, error) {
	var buf bytes.Buffer
	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   []map[string]interface{}{},
				"filter": []map[string]interface{}{},
			},
		},
		"from": offset,
		"size": limit,
	}

	// Thêm query (tìm kiếm theo tên hoặc mô tả)
	if query != "" {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"name^3", "description"}, // Boost name field
				},
			},
		)
	} else {
		// If no query, match all documents.
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{"match_all": map[string]interface{}{}},
		)
	}

	// Thêm filter theo category_id
	if categoryID != "" {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{"term": map[string]interface{}{"category_id": categoryID}},
		)
	}

	// Thêm filter theo khoảng giá
	priceRange := map[string]interface{}{}
	if minPrice > 0 {
		priceRange["gte"] = minPrice
	}
	if maxPrice > 0 && maxPrice >= minPrice {
		priceRange["lte"] = maxPrice
	}
	if len(priceRange) > 0 {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{"range": map[string]interface{}{"price": priceRange}},
		)
	}

	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		return nil, 0, fmt.Errorf("%w: failed to encode search query: %v", domain.ErrFailedToSearchProducts, err)
	}

	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex(indexName),
		r.es.Search.WithBody(&buf),
		r.es.Search.WithTrackTotalHits(true), // Để lấy tổng số hits chính xác
	)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: failed to perform search request: %v", domain.ErrFailedToSearchProducts, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, 0, fmt.Errorf("%w: failed to parse ES error response: %v", domain.ErrFailedToSearchProducts, err)
		}
		return nil, 0, fmt.Errorf("%w: Elasticsearch search error: [%s] %s", domain.ErrFailedToSearchProducts, res.Status(), e["error"])
	}

	var resMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, 0, fmt.Errorf("%w: failed to decode ES response: %v", domain.ErrFailedToSearchProducts, err)
	}

	var products []*domain.SearchProduct
	totalHits := int64(0)

	if hits, ok := resMap["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				totalHits = int64(value)
			}
		}

		if hitsArr, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsArr {
				if source, ok := hit.(map[string]interface{})["_source"]; ok {
					productJSON, err := json.Marshal(source)
					if err != nil {
						log.Printf("Warning: Failed to marshal product source: %v", err)
						continue
					}
					var p domain.SearchProduct
					if err := json.Unmarshal(productJSON, &p); err != nil {
						log.Printf("Warning: Failed to unmarshal product: %v", err)
						continue
					}
					products = append(products, &p)
				}
			}
		}
	}

	return products, totalHits, nil
}

// DeleteProductFromIndex deletes a product from the search engine by its ID.
// DeleteProductFromIndex xóa một sản phẩm khỏi chỉ mục tìm kiếm bằng ID của nó.
func (r *ElasticsearchProductRepository) DeleteProductFromIndex(ctx context.Context, productID string) error {
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: productID,
		Refresh:    "true", // Refresh the index immediately for testing
	}

	res, err := req.Do(ctx, r.es)
	if err != nil {
		return fmt.Errorf("%w: failed to perform delete request: %v", domain.ErrFailedToDeleteFromIndex, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// Check if it's a 404 Not Found error (product not in index)
		if res.StatusCode == http.StatusNotFound {
			return domain.ErrProductNotFoundInSearch
		}

		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("%w: failed to parse ES error response: %v", domain.ErrFailedToDeleteFromIndex, err)
		}
		return fmt.Errorf("%w: Elasticsearch delete error: [%s] %s", domain.ErrFailedToDeleteFromIndex, res.Status(), e["error"])
	}

	log.Printf("Deleted product %s from index", productID)
	return nil
}
