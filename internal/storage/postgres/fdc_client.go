package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const fdcSearchURL = "https://api.nal.usda.gov/fdc/v1/foods/search"

// USDA FoodData Central nutrient IDs (per 100 g).
const (
	fdcNutrientProtein  = 1003
	fdcNutrientFat      = 1004
	fdcNutrientCarb     = 1005
	fdcNutrientEnergy   = 1008
)

type fdcClient struct {
	apiKey string
	client *http.Client
}

func newFDCClient(apiKey string) *fdcClient {
	return &fdcClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 25 * time.Second},
	}
}

type fdcSearchRequest struct {
	Query    string   `json:"query"`
	PageSize int      `json:"pageSize"`
	DataType []string `json:"dataType"`
}

type fdcSearchResponse struct {
	Foods []fdcFood `json:"foods"`
}

type fdcFood struct {
	Description   string         `json:"description"`
	FoodNutrients []fdcNutrient  `json:"foodNutrients"`
}

type fdcNutrient struct {
	NutrientID int     `json:"nutrientId"`
	Value      float64 `json:"value"`
}

func (c *fdcClient) lookupPer100g(ctx context.Context, query string) (protein, carb, fat, calories float64, ok bool, err error) {
	body, err := json.Marshal(fdcSearchRequest{
		Query:    query,
		PageSize: 1,
		DataType: []string{"Foundation", "SR Legacy"},
	})
	if err != nil {
		return 0, 0, 0, 0, false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fdcSearchURL+"?api_key="+c.apiKey, bytes.NewReader(body))
	if err != nil {
		return 0, 0, 0, 0, false, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return 0, 0, 0, 0, false, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, 0, 0, 0, false, err
	}
	if res.StatusCode != http.StatusOK {
		return 0, 0, 0, 0, false, fmt.Errorf("USDA FDC HTTP %d: %s", res.StatusCode, truncate(string(raw), 200))
	}

	var parsed fdcSearchResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return 0, 0, 0, 0, false, err
	}
	if len(parsed.Foods) == 0 {
		return 0, 0, 0, 0, false, nil
	}

	for _, n := range parsed.Foods[0].FoodNutrients {
		switch n.NutrientID {
		case fdcNutrientProtein:
			protein = n.Value
		case fdcNutrientCarb:
			carb = n.Value
		case fdcNutrientFat:
			fat = n.Value
		case fdcNutrientEnergy:
			calories = n.Value
		}
	}

	if protein == 0 && carb == 0 && fat == 0 && calories == 0 {
		return 0, 0, 0, 0, false, nil
	}
	if calories <= 0 {
		calories = protein*4 + carb*4 + fat*9
	}
	return protein, carb, fat, calories, true, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
