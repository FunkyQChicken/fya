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
	return p.cartCreated
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


type customer struct {
	Email string						`json:"email"`
	Phone string						`json:"phone"`
	Id string							`json:"id"`
	LastName string						`json:"lastName"`
	FirstName string					`json:"firstName"`
	IdentityProvider string				`json:"identityProvider"`
	Loyaltynum string					`json:"loyaltyNum"`
}


type cafe struct {
	Pagernum float64					`json:"pagerNum"`
	Id int							`json:"id"`
	Name string							`json:"name,omitempty"`
	ExternalName string					`json:"externalName,omitempty"`
	Address string						`json:"address,omitempty"`
	City string							`json:"city,omitempty"`
	CountryDivision string				`json:"countryDivision,omitempty"`
	PostalCode string					`json:"postalCode,omitempty"`
	PhoneNumber string					`json:"phoneNumber,omitempty"`
	Cafetz string						`json:"cafeTZ,omitempty"`
	Type string							`json:"type,omitempty"`
	Country string						`json:"country,omitempty"`
	LocationType string					`json:"locationType,omitempty"`
}


type cartsummary struct {
	Status string						`json:"status,omitempty"`
	Destination string					`json:"destination"`
	Priority string						`json:"priority"`
	ClientType string					`json:"clientType"`
	AppVersion string					`json:"appVersion,omitempty"`
	DeliveryFee float32					`json:"deliveryFee"`
	SubTotal float64					`json:"subTotal,omitempty"`
	TaxExempt bool						`json:"taxExempt,omitempty"`
	TotalPrice float64					`json:"totalPrice,omitempty"`
	LeadTime float64					`json:"leadTime"`
	LanguageCode string					`json:"languageCode"`
	SpecialInstructions string			`json:"specialInstructions"`
	OrderStartdt string					`json:"orderStartDT,omitempty"`
	OrderFulfillmentdt string			`json:"orderFulfillmentDT,omitempty"`
	SendToConcur bool					`json:"sendToConcur,omitempty"`
  Tax float64               `json:"tax"`
  Discount float64          `json:"discount"`
}


type cart struct {
  OrderId string          `json:"orderId"` 
	Cartid string						`json:"cartId,omitempty"`
	CreateGroupOrder bool				`json:"createGroupOrder"`
	Customer customer					`json:"customer"`
	ServiceFeeSupported bool			`json:"serviceFeeSupported"`
	Cafes []cafe						`json:"cafes"`
	ApplyDynamicPricing bool			`json:"applyDynamicPricing"`
	SubscriberPricingSupported bool		`json:"subscriberPricingSupported,omitempty"`
	CartSummary cartsummary				`json:"cartSummary"`
	CartStatus string					`json:"cartStatus,omitempty"`
  Items []paneraItem        `json:"items"`
  Discounts []paneraDiscount `json:"discounts"`
}


type cartresp struct {
	Cartid string						`json:"cartId"`
}


type menuversion struct {
	CollectionName string				`json:"collectionName"`
	AggregateVersion string				`json:"aggregateVersion"`
}


type composition struct { }


type itemadd struct {
	MsgKitchen string					`json:"msgKitchen"`
	IsNoSideOption bool					`json:"isNoSideOption"`
	Itemid float64						`json:"itemId"`
	Parentid float64					`json:"parentId"`
	Composition composition				`json:"composition"`
	Portion string						`json:"portion"`
	MsgPreparedFor string				`json:"msgPreparedFor"`
	Quantity float64					`json:"quantity"`
	Type string							`json:"type"`
	Promotional bool					`json:"promotional"`
}


type itemsadd struct {
	Items []itemadd						`json:"items"`
}


type pkid struct {
	CafeId int							`json:"cafeId"`
	LangCode string						`json:"langCode"`
	Versionid string					`json:"versionId"`
	LangVersion string					`json:"langVersion"`
}


type category struct {
	Catid float64						`json:"catId"`
	CatMenuType string					`json:"catMenuType"`
	ImgKey string						`json:"imgKey"`
	I18nName string						`json:"i18nName"`
	I18nNameval string					`json:"i18nNameVal"`
	I18nInfo string						`json:"i18nInfo"`
	LogicalName string					`json:"logicalName"`
	IsNavigable bool					`json:"isNavigable"`
	SortWeight float64					`json:"sortWeight"`
	Placards []float64					`json:"placards"`
	HeroPlacards []float64				`json:"heroPlacards"`
	SubCategories []category			`json:"subCategories"`
}


type combomap struct {
	Comboid float64						`json:"comboId"`
	ComboMenuItemid float64				`json:"comboMenuItemId"`
	PortionMatch bool					`json:"portionMatch"`
	ModSetMatch bool					`json:"modSetMatch"`
}


type nutrient struct {
	Unit string							`json:"unit"`
	Value float64						`json:"value"`
	I18nName string						`json:"i18nName"`
	I18nNameval string					`json:"i18nNameVal"`
	LogicalName string					`json:"logicalName"`
	Nutrient string						`json:"nutrient"`
	Nutrientid float64					`json:"nutrientId"`
	NutrSortWeight float64				`json:"nutrSortWeight"`
}


type allergen struct {
	Allergenid string					`json:"id"`
	Name string							`json:"name"`
	Group string						`json:"group"`
	Risk string							`json:"risk"`
	RiskRanking float64					`json:"riskRanking"`
	IsParent bool						`json:"isParent"`
	I18nName string						`json:"i18nName"`
	I18nNameval string					`json:"i18nNameVal"`
}


type allergens struct {
	Contains []allergen					`json:"contains"`
}


type wellness struct {
	Id string							`json:"id"`
	Wellnessid string					`json:"wellnessId"`
	Name string							`json:"name"`
	I18nName string						`json:"i18nName"`
	I18nNameval string					`json:"i18nNameVal"`
}


type defaultitem struct {
	Itemid float64						`json:"itemId"`
	Qty float64							`json:"qty"`
	Allergens []allergens				`json:"allergens"`
}


type optset struct {
	SortWeight float64					`json:"sortWeight"`
	SortWeightMobile float64			`json:"sortWeightMobile"`
	SortWeightOmni float64				`json:"sortWeightOmni"`
	Itemid int							`json:"itemId"`
	LogicalName string					`json:"logicalName"`
	I18nName string						`json:"i18nName"`
	I18nNameval string					`json:"i18nNameVal"`
	MyPaneraExclusive bool				`json:"myPaneraExclusive"`
	I18nbtnlbl string					`json:"i18nBtnLbl"`
	I18nbtnlblval string				`json:"i18nBtnLblVal"`
	Portioni18n string					`json:"portionI18n"`
	Portion string						`json:"portion"`
	Portioni18nval string				`json:"portionI18nVal"`
	ItemContext string					`json:"itemContext"`
	EntreeToComboMap []combomap			`json:"entreeToComboMap"`
	ImgKey string						`json:"imgKey"`
	IsDefault bool						`json:"isDefault"`
	IsCustomizable bool					`json:"isCustomizable"`
	HasSyrupModifiers bool				`json:"hasSyrupModifiers"`
	HasCustomizations bool				`json:"hasCustomizations"`
	AllowSpecialinstr bool				`json:"allowSpecialInstr"`
	HasIngredientCustomizations bool	`json:"hasIngredientCustomizations"`
	Price float64						`json:"price"`
	Nutr []nutrient						`json:"nutr"`
	Allergens []allergens				`json:"allergens"`
	Wellness []wellness					`json:"wellness"`
	HighSodiumFlag bool					`json:"highSodiumFlag"`
	I18ningstmnt string					`json:"i18nIngStmnt"`
	DefaultItems []defaultitem			`json:"defaultItems"`
	Ingstmnt string						`json:"ingStmnt"`
}


type defaultside struct {
	Calories nutrient					`json:"calories"`
	LogicalName string					`json:"logicalName"`
	I18nName string						`json:"i18nName"`
	I18nNameval string					`json:"i18nNameVal"`
	Itemid float64						`json:"itmeId"`
	Allergens []allergens				`json:"allergens"`
}


type placard struct {
	Plcid float64						`json:"plcId"`
	ImgKey string						`json:"imgKey"`
	I18nName string						`json:"i18nName"`
	I18ndesc string						`json:"i18nDesc"`
	IsOrderable bool					`json:"isOrderable"`
	HasCustomizations bool				`json:"hasCustomizations"`
	AllowSpecialinstr bool				`json:"allowSpecialInstr"`
	IsCustomizable bool					`json:"isCustomizable"`
	DefaultSide defaultside				`json:"defaultSide"`
	OptSets []optset					`json:"optSets"`
}


type headermsg struct {
	Pick1i18n string					`json:"pick1I18n"`
	Pick2i18n string					`json:"pick2I18n"`
}


type sideitem struct {
	Itemid float64						`json:"itemId"`
	LogicalName string					`json:"logicalName"`
	Price float64						`json:"price"`
	ImgKey string						`json:"imgKey"`
	I18nName string						`json:"i18nName"`
	IsNoSideOption bool					`json:"isNoSideOption"`
	SortWeight float64					`json:"sortWeight"`
	Nutrients []nutrient				`json:"nutrients"`
	Allergens []allergens				`json:"allergens"`
}


type sides struct {
	DefaultSide float64					`json:"defaultSide"`
	I18nName string						`json:"i18nName"`
	SideItems []sideitem				`json:"sideItems"`
}


type combo struct {
	Comboid float64						`json:"comboId"`
	LogicalName string					`json:"logicalName"`
	I18nName string						`json:"i18nName"`
	I18ndesc string						`json:"i18nDesc"`
	ImgKey string						`json:"imgKey"`
	Placards []float64					`json:"placards"`
	MaxAllowed float64					`json:"maxAllowed"`
	MinAllowed float64					`json:"minAllowed"`
	Itemid float64						`json:"itemId"`
	SidesAllowed float64				`json:"sidesAllowed"`
	NextComboid float64					`json:"nextComboId"`
	CombocatHeadermsg headermsg			`json:"comboCatHeaderMsg"`
	Sides sides							`json:"sides"`
	Price float64						`json:"price"`
	Categories []combo					`json:"categories"`
	SubCategories []combo				`json:"subCategories"`
}


type dayoffset struct {
	FromOffset float64					`json:"fromOffset"`
	ToOffset float64					`json:"toOffset"`
	QtyLimit float64					`json:"qtyLimit"`
}


type quantityrule struct {
	RuleName string						`json:"ruleName"`
	RuleScope string					`json:"ruleScope"`
	Itemids []float64					`json:"itemIds"`
	RuleeffFromDate string				`json:"ruleEffFromDate"`
	RuleeffToDate string				`json:"ruleEffToDate"`
	I18nMessage string					`json:"i18nMessage"`
	DayOffset []dayoffset				`json:"dayOffset"`
	TranslatedMessage string			`json:"translatedMessage"`
}


type menu struct {
	Pkid pkid							`json:"pkid"`
	MenuUpdated bool					`json:"menuUpdated"`
	MenuType string						`json:"menuType"`
	Categories map[string]category		`json:"categories"`
	Placards map[string]placard			`json:"placards"`
	Combos map[string]combo				`json:"combos"`
	MenuTransition map[string]float64	`json:"menuTransition"`
	QuantityRuleSet []quantityrule		`json:"quantityRuleSet"`
	AllowedAllergens []allergen			`json:"allowedAllergens"`
}

type paneraDiscount struct {
  Disctype string   `json:"type"`
  Name string       `json:"name"`
  PromoCode string `json:"promoCode"`
  Amount float64 `json:"amount"`
  DiscCode  int64 `json:"discCode"`
  Prerequiesite string `json:"prerequiesite"`
  IsSharable bool `json:"isSharable"`
  DiscountApplicationType string `json:"discountApplicationType"`
  RedemptionCode string `json:"redemptionCode"`
  AutoApply bool `json:"autoApply"`
  SwapItemId int `json:"swapItemId"`
}

type paneraItems struct {
  Items []paneraItem `json:"items"`
}

type paneraItem struct {
  IsNoSideOption bool `json:"isNoSideOption"`
  Itemid float32 `json:"itemid"`
  Parentid int `json:"parentid"`
  ShowTaxOnSeparateReceiptFlag int `json:"showTaxOnSeparateReceiptFlag"`
  TaxBit int `json:"taxBit"`
  TaxabilityIndicator string `json:"taxabilityIndicator"`
  SequenceNum int `json:"sequenceNum"`
  ItemId int `json:"itemId"`
  Type string `json:"type"`
  Name string `json:"name"`
  Amount float32 `json:"amount"`
  TotalPrice float32 `json:"totalPrice"`
  Quantity int `json:"quantity"`
  ItemDiscAmt float32 `json:"itemDiscAmt"`
  AfterDiscAmt int `json:"afterDiscAmt"`
  SalesTaxAmount int `json:"salesTaxAmount"`
  MsgPreparedFor string `json:"msgPreparedFor"`
  MsgKitchen string `json:"msgKitchen"`
  Discounts []paneraDiscount `json:"discounts"`
  RenderSource renderSource `json:"renderSource"`
  Unavailable bool `json:"unavailable"`
  StockedOut bool `json:"stockedOut"`
  Promotional bool `json:"promotional"`
  Composition composition `json:"composition"`
  Portion string `json:"portion"`
  Taxes []tax `json:"taxes"`
}

type tax struct {
  TaxBit int `json:"taxBit"`
  Description string `json:"description"`
  Amount int `json:"amount"`
}


type renderSource struct {
  ProductId int `json:"productId"`
  ParentPlacardId int `json:"parentPlacardId"`
  LogicalName string `json:"logicalName"`
  MenuItemType string `json:"menuItemType"`
  Name string `json:"name"`
  Description string `json:"description"`
  I18nName string `json:"i18nName"`
  I18nDesc string `json:"i18nDesc"`
  ImgKey string `json:"imgKey"`
  IsAvailable int `json:"isAvailable"`
  IsOptSet int `json:"isOptSet"`
  Price float32 `json:"price"`
  Portion string `json:"portion"`
  Nutrients []nutrient `json:"nutrients"`
  HasCustomizations bool `json:"hasCustomizations"`
  AllowSpecialInstr bool `json:"allowSpecialInstr"`
  Wellness []wellness `json:"wellness"`
  I18nIngStmnt string `json:"i18nIngStmnt"`
  DetailedIngredients string `json:"detailedIngredients"`
  NutrientSuffix string `json:"nutrientSuffix"`
  // includedInProduct 
  //"availabilityExceptions"
  //"selfServiceSwap"
}


type discountsReq struct {
  Discounts []discountReq `json:"discounts"`
}

type discountReq struct {
  Disctype string  `json:"type"`
  PromoCode string `json:"promoCode"`
}


