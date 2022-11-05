package main


type fauxchain struct {
  restraunts []location
}

func InitFauxChain() fauxchain {
  fauxstraunts := []fauxstraunt {
    { 
      description: "How did this get here???",
      address: "litterally Narnia",
    },
    { 
      description: "This one makes a little more sense",
      address: "the north pole",
    },
  }
  restraunts := make([]location, 2);
  for i, v := range fauxstraunts {
    restraunts[i] = &v
  }
  return fauxchain{restraunts}
}
func (r *fauxchain) LoadCredentials() bool {return false}
func (r *fauxchain) Login(username string, password string) bool {return true}
func (r *fauxchain) Locations() []location {return r.restraunts}


type fauxstraunt struct {
  description string
  address string
  cartCreated bool
  cart []item
  discountOne bool 
  discountTwo bool
}


func (r *fauxstraunt) GetDescription() string {return r.description}
func (r *fauxstraunt) GetAddress() string {return r.address}
func (r *fauxstraunt) CreateCart() {r.cartCreated = true}
func (r *fauxstraunt) Menu() []item {
  return []item {
    {
      name: "Snowcone",
      description: "A tasty and sweet treat for any to eat",
      calories: 230,
      cost: 299,
      id: 0,
    },
    {
      name: "Chocolate",
      description: "A classic candy that was initially a spicy beverage",
      calories: 150,
      cost: 100,
      id: 1,
    },
    {
      name: "Water",
      description: "We are legally required to provide this to you",
      calories: 0,
      cost: 0,
      id: 2,
    },
  }
}

func (r *fauxstraunt) AddItem(it item) {
  if ! r.cartCreated {
    panic("Cart not created and item added")
  }
  r.cart = append(r.cart, it)
}

func (r *fauxstraunt) Discounts() []discount {
  return []discount {
    {
      name: "subscriber plus",
      description: "five cents off any order!",
      id: 1,
    },
    {
      name: "subscriber minus",
      description: "five cents added to any order!?!?",
      id: 2,
    },
  }
}

func (r *fauxstraunt) ApplyDiscounts(disc discount) {
  if ! r.cartCreated {
    panic("Cart not created and discount applied")
  }
  switch disc.id {
    case 1:
    r.discountOne = true
    case 2:
    r.discountTwo = true
  }
}

func (r *fauxstraunt) Cart() []cartItem {
  ret := make([]cartItem, 0, len(r.cart) + 3)
  for _, it := range r.cart {
    ret = append(ret, cartItem{description: it.name, cost: it.cost,})
  }
  ret = append(ret, cartItem{description: "Tax", cost: 10,})
  if r.discountOne {
    ret = append(ret, cartItem{description: "subscriber plus!", cost: -5,})
  }
  if r.discountTwo {
    ret = append(ret, cartItem{description: "subscriber minus?", cost: 5,})
  }
  return ret
}

func (r *fauxstraunt) Checkout() bool {return r.cartCreated;}

