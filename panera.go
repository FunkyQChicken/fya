package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Panera struct {
	id int
	description string
	address string
	destinationCode string
	credentialsLoaded bool
	cartCreated bool
	cartid string
	cart []item
}

func (p *Panera) GetDescription() string {
	return p.description
}

func (p *Panera) GetAddress() string {
	return p.address
}

func (p *Panera) CreateCart() {
	if !p.credentialsLoaded {
		log.Fatalln("Can’t create cart if credentials haven’t yet been loaded!")
	}
	
	var c cart = cart {
		CreateGroupOrder: false,
		Customer: customer {
			Email: os.Getenv("Email"),
			Phone: os.Getenv("Phone"),
			Id: os.Getenv("Id"),
			LastName: os.Getenv("LastName"),
			FirstName: os.Getenv("FirstName"),
			IdentityProvider: os.Getenv("IdentityProvider"),
			Loyaltynum: os.Getenv("Loyaltynum"),
		},
		ServiceFeeSupported: true,
		Cafes: []cafe {
			{
				Pagernum: 0,
				Id: p.id,
			},
		},
		ApplyDynamicPricing: true,
		CartSummary: cartsummary {
			Destination: p.destinationCode,
			Priority: "ASAP",
			ClientType: "MOBILE_IOS",
			DeliveryFee: 0.0,
			LeadTime: 10,
			LanguageCode: "en-US",
		},
	}

  cr := postRequest[cart, cartresp](
    p.URL("/cart"),
    p.Header(),
    c,
  )

	p.cartid = cr.Cartid
	
	p.cartCreated = true
}

func (p *Panera) Menu() []item {
	if !p.credentialsLoaded {
		log.Fatalln("Can’t construct menu if credentials haven’t yet been loaded!")
	}

  mv := getRequest[menuversion](
    p.URL(fmt.Sprintf("/%d/menu/version", p.id)),
    p.Header(),
  )

  m := getRequest[menu](
		p.URL(fmt.Sprintf("/en-US/%d/menu/v2/%s", p.id, mv.AggregateVersion)),
    p.Header(),
  )
	
  ret := make([]item, 0, 100)
  for _, placard := range m.Placards {
    optsets := placard.OptSets
    if (optsets != nil) {
      for _, optset := range optsets {
        name := optset.I18nName
        description := optset.LogicalName
        calories := 0
        cost := int(optset.Price * 100)
        id := optset.Itemid 
        for _, nutr := range optset.Nutr {
          if nutr.LogicalName ==  "Calories" {
            calories = int(nutr.Value)
          }
        }

        ret = append(ret, item{
          name,
          description,
          calories,
          cost,
          id,
        })
      }
    }
  }
  
	return ret
}

func (p *Panera) AddItem(i item) {
	if !p.cartCreated {
		panic("Item added without an existing cart!")
	}
	
	var isa = itemsadd {
		Items: []itemadd {
			{
				IsNoSideOption: false,
				Itemid: float64(i.id),
				Parentid: 0,
				Composition: composition { },
				MsgPreparedFor: os.Getenv("FirstName"),
				Quantity: 1,
				Type: "PRODUCT",
				Promotional: false,
			},
		},
	}

  postRequestNoMarshal(
	  p.URL(fmt.Sprintf("/v2/cart/%s/item", p.cartid) ), // hupsell=NONE&groupHost=false),
    p.Header(),
    isa)

	p.cart = append(p.cart, i)
}

func (p *Panera) Discounts() []discount {
	return []discount {
		{
      name: "Panera Sip-Club",
      description: "One free drink with refills every 2 hours!",
      id:1238,
    },
	}
}

func (p *Panera) ApplyDiscounts(d discount) {
	if !p.cartCreated {
		panic("Item applied without an existing cart!")
	}
  discBody := discountsReq {
    Discounts: []discountReq {
        {
          Disctype: "WALLET_CODE",
          PromoCode: fmt.Sprintf("%d", d.id),
        },
      },
  } 
  postRequestNoMarshal(
		p.URL(fmt.Sprintf("/cart/%s/discount", p.cartid)),
		p.Header(),
    discBody)
}

func (p *Panera) Cart() []cartItem {
	var cis []cartItem
	cis = make([]cartItem, 0, len(p.cart) + 2) 

  cart := postRequest[struct{}, cart](
		p.URL(fmt.Sprintf("/cart/%s/checkout", p.cartid)), //?summary=true
    p.Header(),
    struct{}{})
    
  for _, it := range cart.Items {
    cost := int(it.Amount * 100)
    name := it.RenderSource.Name 
    cis = append(cis, cartItem{name, cost})
  }

  cis = append(cis, cartItem{"Tax", int( 100 * cart.CartSummary.Tax)})
  cis = append(cis, cartItem{"Discount", int(-100 * cart.CartSummary.Discount)})

  return cis
}

func (p *Panera) Checkout() bool {
	// TODO: Actually check out
  resp := postRequestNoMarshal(
    p.URL(fmt.Sprintf("/payment/v2/slot-submit/%s", p.cartid)),
    p.Header(),
    checkoutReq {})
	return resp.StatusCode == 200
}

func (p *Panera) URL(path string) *url.URL {
	return &url.URL {
		Scheme: "https",
		Host: "services-mob.panerabread.com",
		Path: path,
	}
}

func (p *Panera) Header() map[string][]string {
	return map[string][]string {
		"auth_token": {
			os.Getenv("auth_token"),
		},
		"appVersion": {
			"4.71.0",
		},
		"api_token": {
			os.Getenv("api_token"),
		},
		"Content-Type": {
			"application/json",
		},
		"deviceId": {
			os.Getenv("deviceId"),
		},
		"User-Agent": {
			"Panera/4.69.9 (iPhone; iOS 16.2; Scale/3.00)",
		},
	}
}


type PaneraChain struct {
	name string
	restaurants []*Panera
}

func InitPaneraChain() PaneraChain {
	return PaneraChain {
		name: "Panera",
		restaurants: []*Panera {
			&(Panera {
				id: 203162,
				description: "Rensselaer Union",
				address: "110 8th Street\nTroy, NY 12180",
				destinationCode: "RPU",
			}),
		},
	}
}

func (pc *PaneraChain) GetName() string {
	return pc.name
}

func (pc *PaneraChain) LoadCredentials() bool {
	var err error
	err = godotenv.Load("panera.env")
  todoHandleErrorBetter(err)
  if err != nil {
    return false
  }
	
	var p *Panera
	for _, p = range pc.restaurants {
		p.credentialsLoaded = true
	}
	
	return true
}

func (pc *PaneraChain) Login(username string, password string) bool {
	return true
}

func (pc *PaneraChain) Locations() []location {
	var ls []location = make([]location, 0, len(pc.restaurants))
	var p *Panera
	for _, p = range pc.restaurants {
		ls = append(ls, p)
	}
	return ls
}

