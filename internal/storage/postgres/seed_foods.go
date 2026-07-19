package postgres

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

// foodSeedItem: Name = hiển thị trong app; Query = từ khóa USDA FDC.
// Fallback macro/100g từ USDA public domain (dùng khi không có API key hoặc API lỗi).
type foodSeedItem struct {
	Name     string
	Query    string
	Protein  float64
	Carb     float64
	Fat      float64
	Calories float64
	ServingG float64 // 0 = default 100g
}

var seedFoodsList = []foodSeedItem{
	// Protein — poultry & meat
	{Name: "Chicken breast (cooked)", Query: "chicken breast cooked", Protein: 31.0, Carb: 0, Fat: 3.6, Calories: 165},
	{Name: "Chicken thigh (cooked)", Query: "chicken thigh cooked", Protein: 26.0, Carb: 0, Fat: 10.9, Calories: 209},
	{Name: "Turkey breast (cooked)", Query: "turkey breast cooked", Protein: 30.1, Carb: 0, Fat: 0.7, Calories: 135},
	{Name: "Lean beef (95% lean)", Query: "beef ground 95 lean cooked", Protein: 26.1, Carb: 0, Fat: 5.0, Calories: 151},
	{Name: "Beef sirloin (cooked)", Query: "beef sirloin cooked", Protein: 29.0, Carb: 0, Fat: 9.0, Calories: 201},
	{Name: "Pork tenderloin (cooked)", Query: "pork tenderloin cooked", Protein: 26.0, Carb: 0, Fat: 4.0, Calories: 143},
	{Name: "Bacon (cooked)", Query: "bacon cooked", Protein: 37.0, Carb: 1.4, Fat: 41.8, Calories: 541},
	{Name: "Ham (lean)", Query: "ham lean", Protein: 21.0, Carb: 1.5, Fat: 5.5, Calories: 145},
	{Name: "Lamb (leg, cooked)", Query: "lamb leg cooked", Protein: 28.2, Carb: 0, Fat: 8.0, Calories: 191},

	// Protein — fish & seafood
	{Name: "Salmon (Atlantic)", Query: "salmon atlantic cooked", Protein: 25.4, Carb: 0, Fat: 13.4, Calories: 206},
	{Name: "Tuna (canned in water)", Query: "tuna canned water", Protein: 23.6, Carb: 0, Fat: 0.8, Calories: 109},
	{Name: "Cod (cooked)", Query: "cod cooked", Protein: 23.0, Carb: 0, Fat: 0.9, Calories: 105},
	{Name: "Tilapia (cooked)", Query: "tilapia cooked", Protein: 26.2, Carb: 0, Fat: 2.7, Calories: 128},
	{Name: "Shrimp (cooked)", Query: "shrimp cooked", Protein: 24.0, Carb: 0.2, Fat: 0.3, Calories: 99},
	{Name: "Crab (cooked)", Query: "crab cooked", Protein: 20.5, Carb: 0, Fat: 1.5, Calories: 97},
	{Name: "Sardines (canned)", Query: "sardines canned", Protein: 24.6, Carb: 0, Fat: 11.5, Calories: 208},

	// Protein — eggs & dairy (serving = 1 large egg ~50g unless noted)
	{Name: "Egg (whole)", Query: "egg whole raw", Protein: 6.3, Carb: 0.35, Fat: 5.3, Calories: 78, ServingG: 50},
	{Name: "Egg white", Query: "egg white", Protein: 3.6, Carb: 0.24, Fat: 0.07, Calories: 17, ServingG: 33},
	{Name: "Greek yogurt (plain, nonfat)", Query: "yogurt greek plain nonfat", Protein: 10.2, Carb: 3.6, Fat: 0.4, Calories: 59},
	{Name: "Greek yogurt (plain, 2%)", Query: "yogurt greek plain 2 percent", Protein: 9.0, Carb: 3.9, Fat: 2.0, Calories: 73},
	{Name: "Cottage cheese (low fat)", Query: "cottage cheese low fat", Protein: 11.1, Carb: 3.4, Fat: 1.0, Calories: 72},
	{Name: "Cheddar cheese", Query: "cheddar cheese", Protein: 24.9, Carb: 1.3, Fat: 33.1, Calories: 403},
	{Name: "Mozzarella (part skim)", Query: "mozzarella part skim", Protein: 24.3, Carb: 2.8, Fat: 15.9, Calories: 254},
	{Name: "Milk (skim)", Query: "milk skim", Protein: 3.4, Carb: 5.0, Fat: 0.1, Calories: 34},
	{Name: "Milk (2%)", Query: "milk 2 percent", Protein: 3.4, Carb: 4.8, Fat: 2.0, Calories: 50},
	{Name: "Milk (whole)", Query: "milk whole", Protein: 3.2, Carb: 4.8, Fat: 3.3, Calories: 61},

	// Protein — plant & supplements
	{Name: "Tofu (firm)", Query: "tofu firm", Protein: 17.3, Carb: 2.8, Fat: 8.7, Calories: 144},
	{Name: "Tempeh", Query: "tempeh", Protein: 19.0, Carb: 9.4, Fat: 10.8, Calories: 192},
	{Name: "Edamame (cooked)", Query: "edamame cooked", Protein: 11.9, Carb: 8.9, Fat: 5.2, Calories: 122},
	{Name: "Whey protein powder", Query: "whey protein powder", Protein: 75.0, Carb: 8.0, Fat: 5.0, Calories: 370},
	{Name: "Casein protein powder", Query: "casein protein powder", Protein: 75.0, Carb: 6.0, Fat: 3.0, Calories: 350},
	{Name: "Pea protein powder", Query: "pea protein powder", Protein: 80.0, Carb: 4.0, Fat: 6.0, Calories: 380},

	// Carbs — grains & starches
	{Name: "White rice (cooked)", Query: "rice white cooked", Protein: 2.7, Carb: 28.2, Fat: 0.3, Calories: 130},
	{Name: "Brown rice (cooked)", Query: "rice brown cooked", Protein: 2.6, Carb: 23.0, Fat: 0.9, Calories: 112},
	{Name: "Jasmine rice (cooked)", Query: "rice cooked", Protein: 2.7, Carb: 28.0, Fat: 0.3, Calories: 129},
	{Name: "Quinoa (cooked)", Query: "quinoa cooked", Protein: 4.4, Carb: 21.3, Fat: 1.9, Calories: 120},
	{Name: "Oats (dry)", Query: "oats dry", Protein: 13.2, Carb: 67.7, Fat: 6.5, Calories: 379},
	{Name: "Oatmeal (cooked)", Query: "oatmeal cooked", Protein: 2.5, Carb: 12.0, Fat: 1.4, Calories: 71},
	{Name: "Couscous (cooked)", Query: "couscous cooked", Protein: 3.8, Carb: 23.2, Fat: 0.2, Calories: 112},
	{Name: "Pasta (cooked)", Query: "pasta cooked", Protein: 5.8, Carb: 30.9, Fat: 0.9, Calories: 157},
	{Name: "Rice noodles (cooked)", Query: "rice noodles cooked", Protein: 1.8, Carb: 25.0, Fat: 0.2, Calories: 109},
	{Name: "Bread (white)", Query: "bread white", Protein: 9.0, Carb: 49.0, Fat: 3.2, Calories: 266},
	{Name: "Bread (whole wheat)", Query: "bread whole wheat", Protein: 13.0, Carb: 41.0, Fat: 4.2, Calories: 252},
	{Name: "Bagel (plain)", Query: "bagel plain", Protein: 10.5, Carb: 53.0, Fat: 1.5, Calories: 257},
	{Name: "Tortilla (flour)", Query: "tortilla flour", Protein: 8.0, Carb: 49.0, Fat: 8.0, Calories: 304},
	{Name: "Rice cake (plain)", Query: "rice cake plain", Protein: 7.0, Carb: 81.0, Fat: 2.8, Calories: 387},

	// Carbs — vegetables & legumes
	{Name: "Sweet potato (baked)", Query: "sweet potato baked", Protein: 2.0, Carb: 20.7, Fat: 0.1, Calories: 92},
	{Name: "Potato (baked)", Query: "potato baked flesh", Protein: 2.5, Carb: 21.0, Fat: 0.1, Calories: 93},
	{Name: "Corn (cooked)", Query: "corn sweet cooked", Protein: 3.3, Carb: 21.0, Fat: 1.2, Calories: 96},
	{Name: "Lentils (cooked)", Query: "lentils cooked", Protein: 9.0, Carb: 20.0, Fat: 0.4, Calories: 116},
	{Name: "Chickpeas (cooked)", Query: "chickpeas cooked", Protein: 8.9, Carb: 27.4, Fat: 2.6, Calories: 164},
	{Name: "Black beans (cooked)", Query: "black beans cooked", Protein: 8.9, Carb: 23.7, Fat: 0.5, Calories: 132},
	{Name: "Kidney beans (cooked)", Query: "kidney beans cooked", Protein: 8.7, Carb: 22.8, Fat: 0.5, Calories: 127},

	// Fruits
	{Name: "Banana", Query: "banana raw", Protein: 1.1, Carb: 22.8, Fat: 0.3, Calories: 89},
	{Name: "Apple", Query: "apple raw", Protein: 0.3, Carb: 13.8, Fat: 0.2, Calories: 52},
	{Name: "Orange", Query: "orange raw", Protein: 0.9, Carb: 11.8, Fat: 0.1, Calories: 47},
	{Name: "Blueberries", Query: "blueberries raw", Protein: 0.7, Carb: 14.5, Fat: 0.3, Calories: 57},
	{Name: "Strawberries", Query: "strawberries raw", Protein: 0.7, Carb: 7.7, Fat: 0.3, Calories: 32},
	{Name: "Grapes", Query: "grapes raw", Protein: 0.7, Carb: 18.1, Fat: 0.2, Calories: 69},
	{Name: "Mango", Query: "mango raw", Protein: 0.8, Carb: 15.0, Fat: 0.4, Calories: 60},
	{Name: "Watermelon", Query: "watermelon raw", Protein: 0.6, Carb: 7.6, Fat: 0.2, Calories: 30},
	{Name: "Dates (dried)", Query: "dates medjool", Protein: 1.8, Carb: 75.0, Fat: 0.2, Calories: 277},

	// Vegetables
	{Name: "Broccoli (cooked)", Query: "broccoli cooked", Protein: 2.4, Carb: 7.2, Fat: 0.4, Calories: 35},
	{Name: "Spinach (raw)", Query: "spinach raw", Protein: 2.9, Carb: 3.6, Fat: 0.4, Calories: 23},
	{Name: "Kale (raw)", Query: "kale raw", Protein: 2.9, Carb: 4.4, Fat: 0.9, Calories: 35},
	{Name: "Carrot (raw)", Query: "carrot raw", Protein: 0.9, Carb: 9.6, Fat: 0.2, Calories: 41},
	{Name: "Bell pepper (raw)", Query: "pepper sweet raw", Protein: 1.0, Carb: 6.0, Fat: 0.3, Calories: 26},
	{Name: "Tomato (raw)", Query: "tomato raw", Protein: 0.9, Carb: 3.9, Fat: 0.2, Calories: 18},
	{Name: "Cucumber (raw)", Query: "cucumber raw", Protein: 0.7, Carb: 3.6, Fat: 0.1, Calories: 15},
	{Name: "Mushrooms (white)", Query: "mushrooms white raw", Protein: 3.1, Carb: 3.3, Fat: 0.3, Calories: 22},
	{Name: "Asparagus (cooked)", Query: "asparagus cooked", Protein: 2.4, Carb: 4.1, Fat: 0.2, Calories: 22},
	{Name: "Green beans (cooked)", Query: "green beans cooked", Protein: 1.9, Carb: 7.9, Fat: 0.2, Calories: 35},
	{Name: "Cauliflower (cooked)", Query: "cauliflower cooked", Protein: 1.8, Carb: 4.1, Fat: 0.5, Calories: 23},
	{Name: "Zucchini (cooked)", Query: "zucchini cooked", Protein: 1.2, Carb: 3.5, Fat: 0.4, Calories: 17},
	{Name: "Bok choy (cooked)", Query: "bok choy cooked", Protein: 1.6, Carb: 2.2, Fat: 0.2, Calories: 12},

	// Fats & nuts
	{Name: "Avocado", Query: "avocado raw", Protein: 2.0, Carb: 8.5, Fat: 14.7, Calories: 160},
	{Name: "Almonds", Query: "almonds", Protein: 21.2, Carb: 21.6, Fat: 49.9, Calories: 579},
	{Name: "Walnuts", Query: "walnuts", Protein: 15.2, Carb: 13.7, Fat: 65.2, Calories: 654},
	{Name: "Cashews", Query: "cashews", Protein: 18.2, Carb: 30.2, Fat: 43.8, Calories: 553},
	{Name: "Peanuts", Query: "peanuts", Protein: 25.8, Carb: 16.1, Fat: 49.2, Calories: 567},
	{Name: "Peanut butter", Query: "peanut butter", Protein: 22.5, Carb: 22.3, Fat: 49.9, Calories: 597},
	{Name: "Almond butter", Query: "almond butter", Protein: 21.0, Carb: 18.8, Fat: 55.5, Calories: 614},
	{Name: "Chia seeds", Query: "chia seeds", Protein: 16.5, Carb: 42.1, Fat: 30.7, Calories: 486},
	{Name: "Flax seeds", Query: "flax seeds", Protein: 18.3, Carb: 28.9, Fat: 42.2, Calories: 534},
	{Name: "Olive oil", Query: "olive oil", Protein: 0, Carb: 0, Fat: 100, Calories: 884},
	{Name: "Coconut oil", Query: "coconut oil", Protein: 0, Carb: 0, Fat: 100, Calories: 862},
	{Name: "Butter", Query: "butter salted", Protein: 0.9, Carb: 0.1, Fat: 81.1, Calories: 717},

	// Misc & snacks
	{Name: "Honey", Query: "honey", Protein: 0.3, Carb: 82.4, Fat: 0, Calories: 304},
	{Name: "Dark chocolate (70%)", Query: "chocolate dark 70", Protein: 7.8, Carb: 45.9, Fat: 42.6, Calories: 598},
	{Name: "Protein bar (average)", Query: "protein bar", Protein: 30.0, Carb: 35.0, Fat: 10.0, Calories: 350},
	{Name: "Hummus", Query: "hummus", Protein: 7.9, Carb: 14.9, Fat: 9.6, Calories: 166},
	{Name: "Salsa", Query: "salsa ready-to-serve", Protein: 1.5, Carb: 7.0, Fat: 0.2, Calories: 36},
}

func seedFoods(db *gorm.DB) error {
	apiKey := strings.TrimSpace(os.Getenv("USDA_FDC_API_KEY"))
	var fdc *fdcClient
	if apiKey != "" {
		fdc = newFDCClient(apiKey)
		log.Printf("seed foods: USDA FoodData Central API (key configured)")
	} else {
		log.Printf("seed foods: no USDA_FDC_API_KEY — using embedded macros (USDA public domain)")
	}

	ctx := context.Background()
	created := 0
	fromAPI := 0

	for _, item := range seedFoodsList {
		var n int64
		if err := db.Model(&entity.Food{}).Where("name = ?", item.Name).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}

		protein := item.Protein
		carb := item.Carb
		fat := item.Fat
		calories := item.Calories

		if fdc != nil {
			p, c, f, cal, ok, err := fdc.lookupPer100g(ctx, item.Query)
			if err != nil {
				log.Printf("seed foods: API %q failed (%v), using fallback", item.Query, err)
			} else if ok {
				protein, carb, fat, calories = p, c, f, cal
				fromAPI++
			} else {
				log.Printf("seed foods: no API result for %q, using fallback", item.Query)
			}
			time.Sleep(400 * time.Millisecond)
		}

		if calories <= 0 {
			calories = protein*4 + carb*4 + fat*9
		}

		servingG := item.ServingG
		if servingG <= 0 {
			servingG = 100
		}
		if servingG != 100 {
			scale := servingG / 100
			protein = round2(protein * scale)
			carb = round2(carb * scale)
			fat = round2(fat * scale)
			calories = round2(calories * scale)
		}

		row := &entity.Food{
			Name:         item.Name,
			Protein:      round2(protein),
			Carb:         round2(carb),
			Fat:          round2(fat),
			Calories:     round2(calories),
			ServingSizeG: servingG,
		}
		if err := db.Create(row).Error; err != nil {
			return err
		}
		created++
	}

	if created > 0 {
		if fdc != nil {
			log.Printf("seed: added %d foods (%d from USDA API, %d fallback)", created, fromAPI, created-fromAPI)
		} else {
			log.Printf("seed: added %d foods (embedded macros /100g)", created)
		}
	}
	return nil
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
