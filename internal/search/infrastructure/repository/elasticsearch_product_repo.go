// internal/search/infrastructure/repository/elasticsearch_product_repo.go
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	// "errors"
	"fmt"
	"log" // Thêm log để ghi lỗi

	// "time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi" // Thêm esapi

	// "github.com/elastic/go-elasticsearch/v8/esutil" // Thêm esutil

	"github.com/datngth03/ecommerce-go-app/internal/search/domain" // Import domain package
)

// ElasticsearchProductRepository implements the domain.SearchProductRepository interface.
// ElasticsearchProductRepository triển khai giao diện domain.SearchProductRepository.
type ElasticsearchProductRepository struct {
	es *elasticsearch.Client
}

// NewElasticsearchProductRepository creates a new instance of ElasticsearchProductRepository.
// NewElasticsearchProductRepository tạo một thể hiện mới của ElasticsearchProductRepository.
func NewElasticsearchProductRepository(addr string) (*ElasticsearchProductRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{addr},
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Ping the cluster to check connectivity
	res, err := esClient.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get Elasticsearch info: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("error from Elasticsearch Info API: %s", res.String())
	}
	log.Printf("Connected to Elasticsearch cluster: %s", res.String())

	repo := &ElasticsearchProductRepository{es: esClient}

	// Khởi tạo chỉ mục (index) nếu chưa tồn tại
	if err := repo.createProductIndex(); err != nil {
		return nil, fmt.Errorf("failed to create product index in Elasticsearch: %w", err)
	}

	return repo, nil
}

// createProductIndex creates the Elasticsearch index for products if it doesn't exist.
// createProductIndex tạo chỉ mục Elasticsearch cho sản phẩm nếu chưa tồn tại.
func (r *ElasticsearchProductRepository) createProductIndex() error {
	exists, err := r.indexExists("products")
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}
	if exists {
		log.Println("Elasticsearch index 'products' already exists. Skipping creation.")
		return nil
	}

	// Define the mapping for the 'products' index
	mapping := `
	{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		},
		"mappings": {
			"properties": {
				"id":           {"type": "keyword"},
				"name":         {"type": "text", "analyzer": "standard"},
				"description":  {"type": "text", "analyzer": "standard"},
				"price":        {"type": "float"},
				"category_id":  {"type": "keyword"},
				"image_urls":   {"type": "keyword"},
				"stock_quantity": {"type": "integer"},
				"created_at":   {"type": "date"},
				"updated_at":   {"type": "date"}
			}
		}
	}
	`
	res, err := r.es.Indices.Create("products", r.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch index creation: %s", res.String())
	}
	log.Println("Elasticsearch index 'products' created successfully.")
	return nil
}

// indexExists checks if a given index exists in Elasticsearch.
// indexExists kiểm tra xem một chỉ mục có tồn tại trong Elasticsearch hay không.
func (r *ElasticsearchProductRepository) indexExists(indexName string) (bool, error) {
	res, err := r.es.Indices.Exists([]string{indexName})
	if err != nil {
		return false, fmt.Errorf("failed to check Elasticsearch index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil // Index exists
	}
	if res.StatusCode == 404 {
		return false, nil // Index does not exist
	}

	return false, fmt.Errorf("unexpected status code when checking index existence: %d", res.StatusCode)
}

// IndexProduct implements domain.SearchProductRepository.
// IndexProduct triển khai phương thức IndexProduct của domain.SearchProductRepository.
func (r *ElasticsearchProductRepository) IndexProduct(ctx context.Context, product *domain.SearchProduct) error {
	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrFailedToIndexProduct, err)
	}

	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: product.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true", // Refresh the index immediately for testing/quick visibility
	}

	res, err := req.Do(ctx, r.es)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrFailedToIndexProduct, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("%w: %s", domain.ErrFailedToIndexProduct, res.String())
	}
	return nil
}

// SearchProducts implements domain.SearchProductRepository.
// SearchProducts triển khai phương thức SearchProducts của domain.SearchProductRepository.
func (r *ElasticsearchProductRepository) SearchProducts(ctx context.Context, query string, categoryID string, minPrice float64, maxPrice float64, limit, offset int32) ([]*domain.SearchProduct, int64, error) {
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

	if query != "" {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"name^2", "description"}, // Boost 'name' field
				},
			},
		)
	} else {
		// If no query, match all documents
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{}),
			map[string]interface{}{"match_all": map[string]interface{}{}},
		)
	}

	if categoryID != "" {
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"term": map[string]interface{}{"category_id": categoryID},
			},
		)
	}

	if minPrice > 0 || maxPrice > 0 {
		priceRange := map[string]interface{}{}
		if minPrice > 0 {
			priceRange["gte"] = minPrice
		}
		if maxPrice > 0 {
			priceRange["lte"] = maxPrice
		}
		searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchBody["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{"price": priceRange},
			},
		)
	}

	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToSearchProducts, err)
	}

	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex("products"),
		r.es.Search.WithBody(&buf),
		r.es.Search.WithTrackTotalHits(true), // For getting total count
		r.es.Search.WithPretty(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToSearchProducts, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("%w: %s", domain.ErrFailedToSearchProducts, res.String())
	}

	var resMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, 0, fmt.Errorf("%w: failed to parse search response: %v", domain.ErrFailedToSearchProducts, err)
	}

	var products []*domain.SearchProduct
	var totalHits int64

	if hits, ok := resMap["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				totalHits = int64(value)
			}
		}

		if items, ok := hits["hits"].([]interface{}); ok {
			for _, item := range items {
				if source, ok := item.(map[string]interface{})["_source"].(map[string]interface{}); ok {
					productJSON, err := json.Marshal(source)
					if err != nil {
						log.Printf("Warning: Failed to marshal product source: %v", err)
						continue
					}
					var product domain.SearchProduct
					if err := json.Unmarshal(productJSON, &product); err != nil {
						log.Printf("Warning: Failed to unmarshal product from JSON: %v", err)
						continue
					}
					products = append(products, &product)
				}
			}
		}
	}

	return products, totalHits, nil
}

// DeleteProductFromIndex implements domain.SearchProductRepository.
// DeleteProductFromIndex triển khai phương thức DeleteProductFromIndex của domain.SearchProductRepository.
func (r *ElasticsearchProductRepository) DeleteProductFromIndex(ctx context.Context, productID string) error {
	req := esapi.DeleteRequest{
		Index:      "products",
		DocumentID: productID,
		Refresh:    "true", // Refresh the index immediately for testing/quick visibility
	}

	res, err := req.Do(ctx, r.es)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrFailedToDeleteProductFromIndex, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// If the error is 'not found', it's not a critical error for deletion
		if res.StatusCode == http.StatusNotFound {
			log.Printf("Product %s not found in Elasticsearch index. Skipping deletion.", productID)
			return nil // Consider it successful if not found
		}
		return fmt.Errorf("%w: %s", domain.ErrFailedToDeleteProductFromIndex, res.String())
	}
	return nil
}
