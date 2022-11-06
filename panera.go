package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	
	var err error
	var req http.Request
	var resp *http.Response
	var body []byte
	
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
				Id: fmt.Sprintf("%d", p.id),
			},
		},
		ApplyDynamicPricing: true,
		CartSummary: cartsummary {
			Destination: p.destinationCode,
			Priority: "ASAP",
			ClientType: "MOBILE_IOS",
			DeliveryFee: "0.00",
			LeadTime: 10,
			LanguageCode: "en-US",
		},
	}
	body, err = json.Marshal(c)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	
	req = http.Request {
		Method: "POST",
		URL: p.URL("/cart"),
		Header: p.Header(),
		Body: io.NopCloser(bytes.NewBuffer(body)),
	}
	resp, err = http.DefaultClient.Do(&req)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	
	// var f *os.File
	// f, err = os.OpenFile("panera.log", os.O_CREATE | os.O_RDWR, 0666)
	// defer f.Close()
	// log.SetOutput(f)
	// log.Output(1, string(body))
	// var cartMarshal, _ = json.Marshal(c)
	// log.Output(1, string(cartMarshal))
	
	var cr cartresp
	err = json.Unmarshal(body, &cr)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	p.cartid = cr.Cartid
	
	p.cartCreated = true
}

func (p *Panera) Menu() []item {
	if !p.credentialsLoaded {
		log.Fatalln("Can’t construct menu if credentials haven’t yet been loaded!")
	}
	
	var err error
	var req http.Request
	var resp *http.Response
	var body []byte
	
	// Menu version
	req = http.Request {
		Method: "GET",
		URL: p.URL(fmt.Sprintf("/%d/menu/version", p.id)),
		Header: p.Header(),
	}
	resp, err = http.DefaultClient.Do(&req)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	var mv menuversion
	err = json.Unmarshal(body, &mv)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	
	// Menu
	req = http.Request {
		Method: "GET",
		URL: p.URL(fmt.Sprintf("/en-US/203162/menu/v2/%s", mv.AggregateVersion)),
		Header: p.Header(),
	}
	resp, err = http.DefaultClient.Do(&req)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	var m menu
	err = json.Unmarshal(body, &m)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}

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
	
	var err error
	var req http.Request
	var resp *http.Response
	var body []byte
	
	var isa = itemsadd {
		Items: []itemadd {
			{
				IsNoSideOption: false,
				Itemid: float64(i.id),
				Parentid: 0,
				Composition: composition { },
				MsgPreparedFor: fmt.Sprintf("%s %s", os.Getenv("FirstName"), os.Getenv("LastName")),
				Quantity: 1,
				Type: "PRODUCT",
				Promotional: false,
			},
		},
	}
	body, err = json.Marshal(isa)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	
	var u = p.URL(fmt.Sprintf("/v2/cart/%s/item", p.cartid))
	u.RawQuery = "upsell=NONE&groupHost=false"
	req = http.Request {
		Method: "POST",
		URL: u,
		Header: p.Header(),
		Body: io.NopCloser(bytes.NewBuffer(body)),
	}
	resp, err = http.DefaultClient.Do(&req)
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		// TODO: Handle error more gracefully
		log.Fatalln(err)
	}
	
	var f *os.File
	f, err = os.OpenFile("panera.log", os.O_CREATE | os.O_RDWR, 0666)
	defer f.Close()
	log.SetOutput(f)
	log.Output(1, string(body))
	
	p.cart = append(p.cart, i)
}

func (p *Panera) Discounts() []discount {
	return []discount {
		// TODO: Get discounts
	}
}

func (p *Panera) ApplyDiscounts(d discount) {
	if !p.cartCreated {
		panic("Item applied without an existing cart!")
	}
	// TODO: Apply discounts
}

func (p *Panera) Cart() []cartItem {
	var cis []cartItem
	cis = make([]cartItem, 0, len(p.cart)) // TODO: Add length for additional line items
	var i item
	for _, i = range p.cart {
		cis = append(cis, cartItem {
			description: i.name,
			cost: i.cost,
		})
	}
	// TODO: Finish filling cart
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
	if err != nil {
		// TODO: Handle error more gracefully
		log.Println(err)
		log.Fatalln(err)
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
	Id string							`json:"id"`
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
	DeliveryFee string					`json:"deliveryFee"`
	SubTotal float64					`json:"subTotal,omitempty"`
	TaxExempt bool						`json:"taxExempt,omitempty"`
	TotalPrice float64					`json:"totalPrice,omitempty"`
	LeadTime float64					`json:"leadTime"`
	LanguageCode string					`json:"languageCode"`
	SpecialInstructions string			`json:"specialInstructions"`
	OrderStartdt string					`json:"orderStartDT,omitempty"`
	OrderFulfillmentdt string			`json:"orderFulfillmentDT,omitempty"`
	SendToConcur bool					`json:"sendToConcur,omitempty"`
}


type cart struct {
	Cartid string						`json:"cartId,omitempty"`
	CreateGroupOrder bool				`json:"createGroupOrder"`
	Customer customer					`json:"customer"`
	ServiceFeeSupported bool			`json:"serviceFeeSupported"`
	Cafes []cafe						`json:"cafes"`
	ApplyDynamicPricing bool			`json:"applyDynamicPricing"`
	SubscriberPricingSupported bool		`json:"subscriberPricingSupported,omitempty"`
	CartSummary cartsummary				`json:"cartSummary"`
	CartStatus string					`json:"cartStatus,omitempty"`
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
	Cafeid int							`json:"cafeId"`
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
