package main

type chain interface {
  GetName() string;
  LoginFields() map[string]string;
  Login(fields map[string]string) bool;
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

var fc = InitFauxChain();
var nfc = InitNotFauxChain();
var pc = InitPaneraChain()
var Chains = []chain {
  &fc,
  &nfc,
  pc,
}
