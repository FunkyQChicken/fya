package restaurant

type notFauxchain struct {
}
func InitNotFauxChain() notFauxchain {
  return notFauxchain{}
}
func (r *notFauxchain) GetName() string {return "NotFaux Chain"}
func (r *notFauxchain) LoadCredentials() bool {return false}
func (r *notFauxchain) LoginFields() map[string]string {
  return map[string]string {
    "Username": "faux@straunt.com",
    "Password": "hunter3",
  }
};
func (r *notFauxchain) Login(fields map[string]string) bool {return true};
func (r *notFauxchain) Locations() []Location {
  fauxchain :=  InitFauxChain()
  return (&fauxchain).Locations()
}



type fauxchain struct {
  restaurants []Location
}

func InitFauxChain() fauxchain {
  fauxstraunts := []Location {
    &fauxstraunt { 
      description: "How did this get here???",
      address: "litterally Narnia",
    },
    &fauxstraunt{ 
      description: "This one makes a little more sense",
      address: "the north pole",
    },
  }
  return fauxchain{restaurants: fauxstraunts}
}

func (r *fauxchain) GetName() string {return "FauxChain"}
func (r *fauxchain) LoadCredentials() bool {return false}
func (r *fauxchain) LoginFields() map[string]string {
  return map[string]string {
    "Username": "faux@straunt.com",
    "Password": "hunter2",
  }
};
func (r *fauxchain) Login(fields map[string]string) bool {return true};
func (r *fauxchain) Locations() []Location {return r.restaurants}


type fauxstraunt struct {
  description string
  address string
  cartCreated bool
  cart []FoodItem
  discountOne bool 
  discountTwo bool
}


func (r *fauxstraunt) GetDescription() string {return r.description}
func (r *fauxstraunt) GetAddress() string {return r.address}
func (r *fauxstraunt) CreateCart() {r.cartCreated = true}
func (r *fauxstraunt) Menu() []FoodItem {
  return []FoodItem {
    {
      Name: "Snowcone",
      Description: "A tasty and sweet treat for any to eat",
      Calories: 230,
      Cost: 299,
      Id: 0,
    },
    {
      Name: "Chocolate",
      Description: "A classic candy that was initially a spicy beverage",
      Calories: 150,
      Cost: 100,
      Id: 1,
    },
    {
      Name: "Water",
      Description: "We are legally required to provide this to you",
      Calories: 0,
      Cost: 0,
      Id: 2,
    },
  }
}

func (r *fauxstraunt) AddItem(it FoodItem) {
  if ! r.cartCreated {
    panic("Cart not created and item added")
  }
  r.cart = append(r.cart, it)
}

func (r *fauxstraunt) Discounts() []Discount {
  return []Discount {
    {
      Name: "subscriber plus",
      Description: "five cents off any order!",
      Id: 1,
    },
    {
      Name: "subscriber minus",
      Description: "five cents added to any order!?!?",
      Id: 2,
    },
  }
}

func (r *fauxstraunt) ApplyDiscounts(disc Discount) {
  if ! r.cartCreated {
    panic("Cart not created and discount applied")
  }
  switch disc.Id {
    case 1:
    r.discountOne = true
    case 2:
    r.discountTwo = true
  }
}

func (r *fauxstraunt) Cart() []CartItem {
  ret := make([]CartItem, 0, len(r.cart) + 3)
  for _, it := range r.cart {
    ret = append(ret, CartItem{Description: it.Name, Cost: it.Cost,})
  }
  ret = append(ret, CartItem{Description: "Tax", Cost: 10,})
  if r.discountOne {
    ret = append(ret, CartItem{Description: "subscriber plus!", Cost: -5,})
  }
  if r.discountTwo {
    ret = append(ret, CartItem{Description: "subscriber minus?", Cost: 5,})
  }
  return ret
}

func (r *fauxstraunt) Checkout() bool {return r.cartCreated;}

