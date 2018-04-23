
Announcer
=========
Announcer is an interface for send xml message to online users.

In order to communicate with ejabberd servers it can use client or component connections([XEP-0114](http://xmpp.org/extensions/xep-0114.html)) based on configuration.

This app initialize from [toml](https://github.com/toml-lang/toml) file for standalone and use zookeeper as a runtime configuration initializer and data store for multi node installation.
If you want to run this application standalone you don't need to config zookeeper part.

Configuration
-------------
```toml
rest_api_port=":8080"
debug_port=":6060" #stack trace debuging port

[log]
format="json" #json/text (default:text)
log_level="info" #info/error/warning (default:warning)
log_point="/path/to/log/file" #optional


#Zookeeper connection configuration for datastore and notification center purpose
####this configuration is optional###
[zookeeper]
cluster_nodes="192.168.95.171:2181"
namespace="/watch" #Znode which data will be saved under

#Mysql database connection configuration which contain ws_channel_member table
[mysql]
address="127.0.0.1:3306"
username="root"
password="123"
db="test"

#Redis connection configuration to get online users from
[redis]
cluster_nodes="127.0.0.1:6379"
password=""
db=0
#Time in second duration that connection with redis will be check and if lost try to connect
#In this period of time every users known as online users
check_interval=2 #Second
hash_table="UserPresence" #hash table that online users store in

[ejabberd]
cluster_nodes="192.168.95.180:5222"
default_cluster="A"
rate_limit=123123 #Send message/sec
send_retry=6 #retry number to send a message if sent failed

#Each cluster can ONLY has Client or Component connection
#Please attention to use only one of them
[client]
username="4"
password="4"
domain="soroush.ir"
ping_interval=2

[component]
name="announcer"
secret="announcer"
domain="soroush.ir"
ping_interval=2
```

>To use component connection you must be define component port for ejabberd servers

Rest APIs
---------
|URI|type|Description|
|---|----|-----------|
|/v1/announce/channel|POST|send channel message to that channel online members|
|/v1/announce/user|POST|send a message to user|

Samples
-------

**announce channel message:**

**POST** request to **/v1/announce** with json data like

```json
{
  "channel": 22,
  "message": "<message xml:lang='en' to='%s' type='chat' id='ID_NUMBER' xmlns='jabber:client'><body>MESSAGE_CONTENT</body><body xml:lang='REPLY_ON_THREAD_ID'>989198872580</body><body xml:lang='MAJOR_TYPE'>SIMPLE_CHAT</body><body xml:lang='MINOR_TYPE'>TEXT</body><body xml:lang='REPLY_ON_MESSAGE_ID'>15219732781131af24fc1zwf</body><body xml:lang='SEND_TIME_IN_GMT'>1521973339583</body></message>"
}
```
> **Attention:** You should put ["%s"](https://golang.org/pkg/fmt/) instead of username who this message will be send for

In order to send a message to all online users you need to set channel_id **negative number**.

connection handling
-------------------
**Redis**
Connection to redis server will be check every "check_interval" second.
If we lost our connection we will discard fetch data from that till connection establish again

**Ejabberd Component**
Based on component "ping_interval" configuration; send ping to end server every "ping_interval" second

**Ejabberd Client**
Based on component "ping_interval" configuration; send ping to end server every "ping_interval" second

**Ejabberd connections**
On send time we check that if they are available or not
if lost a connection we will remove that form available connection and when connection numbers goes to 0
we try to connect to all ejabberd nodes again

Debugging
---------
Call `http://localhost:<debug_port>/debug/pprof/trace?seconds=5` to get 5 second of application trace file and then you can see application trace. With
`go tool trace <DOWNLOADED_FILE_PATH>` command you can see what's happen in application on that period of time