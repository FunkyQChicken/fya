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
  GetCustomizations(FoodItem) []FoodOption;
  AddItem(FoodItem, []FoodOption);
  Discounts() []Discount;
  ApplyDiscounts(Discount);
  Cart() []CartItem;
  Checkout() bool;
}

type FoodItem struct {
  Name string
  Description string
  Calories int 
  Cost int // cents
  Id int   // arbitrary, can reference internal array if needed
}


type FoodOption interface {}

type FoodOptionSelectOne struct {
  Name string
  Id int
  Options []string
  Ids []int
  Curr int
}

type FoodOptionsGroup struct {
  Name string
  Id int
  Options []FoodOption
  Selected int
}

type FoodOptionSelectNumber struct {
  Name string
  Id int
  Num int
  Min int
  Max int
}

type FoodOptionText struct {
  Name string
  Text string
  MaxLen int
  DefaultString string
}

type Discount struct {
  Name string
  Description string
  Id int
}

type CartItem struct {
  Description string
  Cost int
}

var fc = InitFauxChain();
var nfc = InitNotFauxChain();
var pc = InitPaneraChain()
var Chains = []Chain {
  &fc,
  &nfc,
  pc,
}
