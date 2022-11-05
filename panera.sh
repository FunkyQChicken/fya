set -e
source panera.env
# provides the following data
# loyalty_num
# user_num
# api_token
# auth_token
# deviceId
# email
# phone
# last_name
# first_name

customer_string='{"email":"'"${email}"'","phone":"'"${phone}"'","id":'"${user_num}"',"lastName":"'"${last_name}"'","firstName":"'"${first_name}"'","identityProvider":"PANERA","loyaltyNum":"'"${loyalty_num}"'"}'
cafe_num='203162' # Rensselaer student union
cafe_string='[{"pagerNum":0,"id":'"${cafe_num}"'}]' 

function pcurl() {
 gum spin --title="${1}" --show-output -- curl -H 'Accept: */*' \
  --no-progress-meter \
  -b "panera.cookies" \
  -c "panera.cookies" \
  -H "auth_token: ${auth_token}" \
  -H 'appVersion: 4.69.9' \
  --compressed -H 'Accept-Language: en-US,en;q=0.9' \
  -H "api_token: ${api_token}" \
  -H 'Content-Type: application/json' \
  -H "deviceId: ${deviceId}" \
  -H 'User-Agent: Panera/4.69.9 (iPhone; iOS 16.2; Scale/3.00)' \
  -H 'Connection: keep-alive' \
  -X "${2}" \
  "${3}" \
  -d "${4}"
}


function get_menu() {
  sip_club_items=`pcurl "Getting sip club items" GET 'https://services-mob.panerabread.com/users/'"${user_num}"'/rewards/v2/'"${loyalty_num}"'/500000/' |
    jq ".rewards | map(select(.name == \"Sip Club - Beverage\")) | .[0].eligibleItems | map(.itemId) | .[]" |
    sed 's/[0-9]\+/"&": true,/g'`

  no_stock=`pcurl "Getting out of stock" GET "https://services-mob.panerabread.com/stockouts?cafeId=203162&busUnit=mobile&fulfillmentDate=2022-11-03" |
    jq ".items | .[]" |
    sed 's/[0-9]\+/"&": true,/g'`

  menu_version=`pcurl "Getting menu version" GET "https://services-mob.panerabread.com/${cafe_num}/menu/version" |
    jq -r .aggregateVersion`


  menu=`pcurl "Getting menu" GET "https://services-mob.panerabread.com/en-US/203162/menu/v2/${menu_version}" |
    jq '.placards | map(.optSets | select(.!= null) | map(select(.itemId | tostring | ({'"${sip_club_items::-1}"'}[.] and null == {'"${no_stock::-1}"'}[.])))) | map(.[])'`

  menu_abridged=`echo $menu | 
    jq -r 'map( (.itemId | tostring) + " ; " + .logicalName) | .[]'`

}

function add_item() {
  drink=`echo -e "${menu_abridged}\nexit" | gum filter`
  if [[ "${drink}" = "exit" ]]; then
    exit
  fi

  id=`echo "${drink}" | awk '{print $1}'`

  directions=`gum input --prompt "Special Instructions? > " --placeholder "Sugar and Cream please"`
  if [[ "" != "$directions" ]]; then
    name=`gum input --prompt "Prepared for? > " --placeholder "Your Name"`
  fi

  request='{"items":[{"msgKitchen":"'"${directions}"'","isNoSideOption":false,"itemId":'${id}',"parentId":0,"composition":{},"portion":"","msgPreparedFor":"'"${name}"'","quantity":1,"type":"PRODUCT","promotional":false}]}'
  add_item=$(pcurl "Adding item" POST "https://services-mob.panerabread.com/v2/cart/${cart_id}/item?upsell=NONE&groupHost=false" "${request}")
}


cart_creation=`pcurl "Creating cart" POST https://services-mob.panerabread.com/cart/ '{"createGroupOrder":false,"customer":'"${customer_string}"',"serviceFeeSupported":true,"cafes":'"${cafe_string}"',"applyDynamicPricing":true,"cartSummary":{"destination":"RPU","priority":"ASAP","clientType":"MOBILE_IOS","deliveryFee":"0.00","leadTime":10,"languageCode":"en-US","specialInstructions":""}}'`
cart_id=`echo $cart_creation | jq .cartId -r`

get_menu
add_item




sip_club=`pcurl "Applying discount" POST "https://services-mob.panerabread.com/cart/${cart_id}/discount" \
  '{"discounts":[{"type":"WALLET_CODE","promoCode":"1238"}]}'`


summary=`pcurl "Getting summary" POST "https://services-mob.panerabread.com/cart/${cart_id}/checkout?summary=true" "{}"`
total_cost=`echo $summary | jq .cartSummary.totalPrice`
sub_total=`echo $summary | jq .cartSummary.subTotal`
tax=`echo $summary | jq .cartSummary.tax`
discount=`echo $summary | jq .cartSummary.discount`
items=`echo $summary | jq -r '.items | map (" * $" + (.amount | tostring) + " :: " +  .renderSource.logicalName) | .[]'`

gum format "
## Panera Cart Checkout
# *Items*
${items}

# *Cost*
* Total Cost: \$${total_cost}
* Cost Calculation: \$${sub_total} + \$${tax} - \$${discount}
 
"
echo

if gum confirm --affirmative="Buy" --negative="Quit" "Are you sure you'd like to place the order?"; then
  res=`pcurl "Placing order" POST "https://services-mob.panerabread.com/payment/v2/slot-submit/${cart_id}" '{"payment":{"giftCards":[],"creditCards":[],"campusCards":[]},"customer":{"smsOptIn":false}}'`
  echo $res
  orderId=`echo $res | jq .orderId`
  gum format "## Order Placed"
else
  gum format "## Order Canceled"
  exit 1
fi

if [[ "$orderId" = "null" ]]; then
  gum format "## Problem With Order"
  echo $res | jq
  exit 1
fi

exit 0

while true; do
  gum format "# Status"
  pcurl "getting status" POST https://services-mob.panerabread.com/orderstatus '{"orderIds":['"${orderId}"']}' | jq
  gum spin --title="Waiting" sleep 10
  sleep 1
done

