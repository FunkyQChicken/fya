package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
)

const paneraCredsFile = "panera_creds.json"

func (p *Panera) GetDescription() string {
	return p.description
}

func (p *Panera) GetAddress() string {
	return p.address
}

func (p *Panera) CreateCart() {
	var c cart = cart {
		CreateGroupOrder: false,
		Customer: customer {
			Email: p.parent.creds.Email,
			Phone: p.parent.creds.Phone,
			Id: p.parent.creds.Id,
			LastName: p.parent.creds.LastName,
			FirstName: p.parent.creds.FirstName,
			IdentityProvider: "PANERA",
			Loyaltynum: p.parent.creds.Loyaltynum,
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
    pURL("/cart"),
    p.parent.creds.authdHeader(),
    c,
  )

	p.cartid = cr.Cartid
	
	p.cartCreated = true
}

func (p *Panera) Menu() []item {
  mv := getRequest[menuversion](
    pURL(fmt.Sprintf("/%d/menu/version", p.id)),
    p.parent.creds.authdHeader(),
  )

  m := getRequest[menu](
		pURL(fmt.Sprintf("/en-US/%d/menu/v2/%s", p.id, mv.AggregateVersion)),
    p.parent.creds.authdHeader(),
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
	  pURL(fmt.Sprintf("/v2/cart/%s/item", p.cartid) ), // hupsell=NONE&groupHost=false),
    p.parent.creds.authdHeader(),
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
		pURL(fmt.Sprintf("/cart/%s/discount", p.cartid)),
    p.parent.creds.authdHeader(),
    discBody)
}

func (p *Panera) Cart() []cartItem {
	var cis []cartItem
	cis = make([]cartItem, 0, len(p.cart) + 2) 

  cart := postRequest[struct{}, cart](
		pURL(fmt.Sprintf("/cart/%s/checkout", p.cartid)), //?summary=true
    p.parent.creds.authdHeader(),
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
    pURL(fmt.Sprintf("/payment/v2/slot-submit/%s", p.cartid)),
    p.parent.creds.authdHeader(),
    checkoutReq {})
	return resp.StatusCode == 200
}

func  pURL(path string) *url.URL {
	return &url.URL {
		Scheme: "https",
		Host: "services-mob.panerabread.com",
		Path: path,
	}
}

func basicHeader() map[string][]string{
  return map[string][]string {
    // Yes, it looks like I've commited an API key
    // Alas, you'd be incorrect
    // It seems that Panera uses one api token for all mobile devices
    "api_token": {
      "bcf0be75-0de6-4af0-be05-13d7470a85f2",
    },
		"appVersion": {
			"4.71.0",
		},
		"Content-Type": {
			"application/json",
		},
		"User-Agent": {
			"Panera/4.69.9 (iPhone; iOS 16.2; Scale/3.00)",
		},
  }
}

func (c *credentials) authdHeader() map[string][]string {
  ret := basicHeader()
  ret["auth_token"] = []string{c.AuthToken}
  //ret["deviceId"] = []string{c.DeviceId}
  return ret
}



func InitPaneraChain() PaneraChain {
	pc := PaneraChain {
		name: "Panera",
	}
  pc.restaurants = []*Panera {
    &(Panera {
      id: 203162,
      description: "Rensselaer Union",
      address: "110 8th Street\nTroy, NY 12180",
      destinationCode: "RPU",
      parent: &pc,
    }),
  }

  return pc
}

func (pc *PaneraChain) GetName() string {
	return pc.name
}

func (pc *PaneraChain) LoadCredentials() bool {
  log.Println("Load Credentials called")
  creds, loaded := tryLoadFromJsonToFile[credentials](paneraCredsFile)
  if (loaded) {
    log.Println("they loaded!")
    pc.creds = creds
    pc.credsLoaded = true
  }
	return loaded
}

func (pc *PaneraChain) Login(token string) bool {
  if pc.credsLoaded {
    log.Fatalln("ERROR: Can't log in again, credentials already loaded")
  }
  // send login request
  rawLoginResp := getRequestNoMarshal(
    pURL(fmt.Sprintf("/token/%s", token)),
    basicHeader(),
  )  
  
  if rawLoginResp.StatusCode != 200 {
    return false
  }
  
  loginResp := responseToJson[tokenResp](rawLoginResp)

  creds := credentials{
    AuthToken: loginResp.AccessToken,
    Email: loginResp.EmailAddress,
    Phone: loginResp.PhoneNumber,
    Id: fmt.Sprint(loginResp.CustomerId),
    FirstName: loginResp.FirstName,
    LastName: loginResp.LastName,
    Loyaltynum: loginResp.Loyalty.CardNumber,
  }

  pc.creds = creds

  saveAsJsonToFile(creds, paneraCredsFile)

  return true
}

func (pc *PaneraChain) Locations() []location {
	if !pc.credsLoaded {
		log.Fatalln("Can’t get locations if credentials haven’t yet been loaded!")
	}
	
	var ls []location = make([]location, 0, len(pc.restaurants))
	var p *Panera
	for _, p = range pc.restaurants {
		ls = append(ls, p)
	}
	return ls
}
