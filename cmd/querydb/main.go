package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	// 打开数据库连接
	db, err := sql.Open("sqlite", "./data/gateway.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 查询提供商
	fmt.Println("=== Providers ===")
	providers, err := queryProviders(db)
	if err != nil {
		log.Printf("Failed to query providers: %v", err)
	} else {
		for _, p := range providers {
			fmt.Printf("ID: %d, Name: %s, Slug: %s, Status: %s\n", p.ID, p.Name, p.Slug, p.Status)
		}
	}

	// 查询模型
	fmt.Println("\n=== Models ===")
	models, err := queryModels(db)
	if err != nil {
		log.Printf("Failed to query models: %v", err)
	} else {
		for _, m := range models {
			fmt.Printf("ID: %d, Provider: %d, Name: %s, Slug: %s, Type: %s\n", 
				m.ID, m.ProviderID, m.Name, m.Slug, m.ModelType)
		}
	}

	// 查询定价
	fmt.Println("\n=== Model Pricing (first 10) ===")
	pricing, err := queryModelPricing(db)
	if err != nil {
		log.Printf("Failed to query model pricing: %v", err)
	} else {
		for i, p := range pricing {
			if i >= 10 { // 只显示前10条
				break
			}
			fmt.Printf("Model: %d, Type: %s, Price: %.8f %s per %s\n", 
				p.ModelID, p.PricingType, p.PricePerUnit, p.Currency, p.Unit)
		}
	}
}

type Provider struct {
	ID     int64
	Name   string
	Slug   string
	Status string
}

type Model struct {
	ID         int64
	ProviderID int64
	Name       string
	Slug       string
	ModelType  string
}

type ModelPricing struct {
	ID           int64
	ModelID      int64
	PricingType  string
	PricePerUnit float64
	Unit         string
	Currency     string
}

func queryProviders(db *sql.DB) ([]Provider, error) {
	query := "SELECT id, name, slug, status FROM providers ORDER BY id"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []Provider
	for rows.Next() {
		var p Provider
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Status)
		if err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}

	return providers, rows.Err()
}

func queryModels(db *sql.DB) ([]Model, error) {
	query := "SELECT id, provider_id, name, slug, model_type FROM models ORDER BY id"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		err := rows.Scan(&m.ID, &m.ProviderID, &m.Name, &m.Slug, &m.ModelType)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}

	return models, rows.Err()
}

func queryModelPricing(db *sql.DB) ([]ModelPricing, error) {
	query := "SELECT id, model_id, pricing_type, price_per_unit, unit, currency FROM model_pricing ORDER BY model_id, pricing_type"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pricing []ModelPricing
	for rows.Next() {
		var p ModelPricing
		err := rows.Scan(&p.ID, &p.ModelID, &p.PricingType, &p.PricePerUnit, &p.Unit, &p.Currency)
		if err != nil {
			return nil, err
		}
		pricing = append(pricing, p)
	}

	return pricing, rows.Err()
}
