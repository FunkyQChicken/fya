package main


type panera struct {
	description string
	address string
	cartCreated bool
	cart []item
}

func (p panera) GetDescription() string {
	return p.description
}

func (p panera) GetAddress() string {
	return p.address
}

func (p *panera) CreateCart() {
	// TODO: Actually create cart
	p.cartCreated = true
}

func (p panera) Menu() []item {
	return []item {
		// TODO: Get menu items
	}
}

func (p *panera) AddItem(i item) {
	if !p.cartCreated {
		panic("Item added without an existing cart!")
	}
	p.cart = append(p.cart, i)
}

func (p panera) Discounts() []discount {
	return []discount {
		// TODO: Get discounts
	}
}

func (p panera) ApplyDiscounts(d discount) {
	if !p.cartCreated {
		panic("Item applied without an existing cart!")
	}
	// TODO: Apply discounts
}

func (p panera) Cart() []cartItem {
	ret := make([]cartItem, 0, len(p.cart)) // TODO: Add length for additional line items
	for _, i := range p.cart {
		ret = append(ret, cartItem{
			description: i.name,
			cost: i.cost,
		})
	}
	// TODO: Finish filling cart
	return ret
}

func (p panera) Checkout() bool {
	// TODO: Actually check out
	return p.cartCreated
}


type panerachain struct {
	restaurants []panera
}

func InitPaneraChain() panerachain {
	return panerachain {
		restaurants: []panera {
			{
				description: "Rensselaer Union",
				address: "110 8th Street\nTroy, NY 12180",
			},
		},
	}
}

func (pc panerachain) LoadCredentials() bool {
	return false
}

func (pc panerachain) Login(username string, password string) bool {
	return true
}

func (pc panerachain) Locations() []panera {
	return pc.restaurants
}
