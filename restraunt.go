package main

type chain interface {
  GetName() string;
  Login(username string, password string) bool;
  LoadCredentials() bool;
  Locations() []location;
}

type location interface {
  GetDescription() string
  GetAddress() string
  CreateCart();
  Menu() []item;
  AddItem(item);
  Discounts() []discount;
  ApplyDiscounts(discount);
  Cart() []cartItem;
  Checkout() bool;
}

type item struct {
  name string
  description string
  calories int 
  cost int // cents
  id int   // arbitrary, can reference internal array if needed
}

type discount struct {
  name string
  description string
  id int
}

type cartItem struct {
  description string
  cost int
}

var fauxchainInstance = InitFauxChain();
var notFauxchainInstance = InitNotFauxChain();
var Chains = []chain {
  &fauxchainInstance,
  &notFauxchainInstance,
}
