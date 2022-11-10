package restaurant

type Chain interface {
  GetName() string;
  LoginFields() map[string]string;
  Login(fields map[string]string) bool;
  LoadCredentials() bool;
  Locations() []Location;
}

type Location interface {
  GetDescription() string
  GetAddress() string
  CreateCart();
  Menu() []FoodItem;
  AddItem(FoodItem);
  Discounts() []Discount;
  ApplyDiscounts(Discount);
  Cart() []CartItem;
  Checkout() bool;
}

type FoodItem struct {
  name string
  description string
  calories int 
  cost int // cents
  id int   // arbitrary, can reference internal array if needed
}

type Discount struct {
  name string
  description string
  id int
}

type CartItem struct {
  description string
  cost int
}

var fc = InitFauxChain();
var nfc = InitNotFauxChain();
var pc = InitPaneraChain()
var Chains = []Chain {
  &fc,
  &nfc,
  pc,
}
