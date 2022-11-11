# FYA
## FYI 

FYA is a TUI to allow ordering from restaurant chains without using their
native websites/applications. 

## Limitations

It is currently in an early version that only supports one panera location on
RPI's campus and only supports checkouts for orders that cost no money (I.E. a
discount was applied).

Additionally, since Panera uses anti-bot cookies for login requests, The
initial login is a little more involved than typing username/passsword. See
below in [Panera's setup section](#panera-setup).


## Setups
### <a name="panera-setup"></a> Setup for Panera

The first time Panera is used you will be prompted for a 'Login response'.
After the first time, the details are saved and loaded from the `~/.fya`
directory.

To get the Login response, First, go to panera's [online
website](https://delivery.panera.ca/orderProcess/). Then, open the 'inspect
element' panel (Ctrl+Shift+c on Firefox, and Ctrl+Shift+i on Chrome) and
navigate to the network tab. Login to panera while the network tab open, and
then search for 'foundation-api/users/uramp' in network tab's filter bar. Right
click on the first request and select copy response. 

This payload has all the information that is needed to login and place orders
using your account. Paste it into the FYA prompt and hit Enter. If this doesn't
work, try reading the log file (`~/.fya/fya.log`) to see what may have gone
wrong.

