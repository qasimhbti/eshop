package main

const (
	promoBOGO = "BOGO"
	promoAPPL = "APPL"
	promoCHMK = "CHMK"
	promoAPOM = "APOM"
)

type cart struct {
	CartItems    []cartItem `json:"cart_items"`
	TotalPrice   float64    `json:"total_price"`
	PromoApplied []string   `json:"promo_applied"`
}

type cartItem struct {
	ProductCode string  `json:"product_code"`
	Price       float64 `json:"price"`
	Promo       string  `json:"promo"`
	Discount    float64 `json:"discount"`
}
type cartManagerImpl struct{}

func (m *cartManagerImpl) calcCartTotAmount(items []*item) *cart {
	hashMapCartItems := make(map[string]bool)
	for _, v := range items {
		item := *v
		if _, ok := hashMapCartItems[item.ProductCode]; !ok {
			hashMapCartItems[item.ProductCode] = true
		}
	}

	cart := &cart{}
	applyPromoCHMK := false
	for _, v := range items {
		item := *v
		ctItem := cartItem{}
		ctItem.ProductCode = item.ProductCode
		ctItem.Price = item.Price
		switch item.ProductCode {
		case "CH1":
			for i := 1; i <= item.Quantity; i++ {
				ctItem.Discount = 0.00
				cart.CartItems = append(cart.CartItems, ctItem)
			}
			cart.TotalPrice += float64(item.Quantity) * (ctItem.Price - ctItem.Discount)

		case "AP1":
			for i := 1; i <= item.Quantity; i++ {
				if item.Quantity >= 3 {
					ctItem.Discount = 1.50
					ctItem.Promo = promoAPPL
					cart.PromoApplied = append(cart.PromoApplied, promoAPPL)
				}
				cart.CartItems = append(cart.CartItems, ctItem)
			}

			appleTotAmount := float64(item.Quantity) * (ctItem.Price - ctItem.Discount)
			if _, ok := hashMapCartItems["OM1"]; ok {
				appleTotAmount /= 2.0
				cart.PromoApplied = append(cart.PromoApplied, promoAPOM)
			}
			cart.TotalPrice += appleTotAmount

		case "CF1":
			for i := 1; i <= item.Quantity; i++ {
				ctItem.Discount = 0.00
				ctItem.Promo = ""
				if i%2 == 0 {
					ctItem.Discount = ctItem.Price
					ctItem.Promo = promoBOGO
					cart.PromoApplied = append(cart.PromoApplied, promoBOGO)
				}
				cart.CartItems = append(cart.CartItems, ctItem)
				cart.TotalPrice += (ctItem.Price - ctItem.Discount)
			}

		case "MK1":
			quantity := item.Quantity
			if _, ok := hashMapCartItems["CH1"]; ok {
				if !applyPromoCHMK {
					applyPromoCHMK = true
					ctItem.Discount = item.Price
					ctItem.Promo = promoCHMK
					cart.CartItems = append(cart.CartItems, ctItem)
					cart.PromoApplied = append(cart.PromoApplied, promoCHMK)
					cart.TotalPrice += (ctItem.Price - ctItem.Discount)
					quantity--
				}
			}
			for i := 1; i <= quantity; i++ {
				ctItem.Discount = 0.00
				ctItem.Promo = ""
				cart.CartItems = append(cart.CartItems, ctItem)
			}
			cart.TotalPrice += float64(quantity) * (ctItem.Price - ctItem.Discount)

		case "OM1":
			for i := 1; i <= item.Quantity; i++ {
				ctItem.Discount = 0.00
				cart.CartItems = append(cart.CartItems, ctItem)
			}
			cart.TotalPrice += float64(item.Quantity) * (ctItem.Price - ctItem.Discount)
		}
	}
	return cart
}
