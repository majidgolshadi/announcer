
Announcer
=========
Announcer is an interface for send xml message to online clients.

In order to communicate with ejabberd servers it use client or component connections([XEP-0114](http://xmpp.org/extensions/xep-0114.html)).

This app initialize from [toml](https://github.com/toml-lang/toml) file and use zookeeper as a runtime configuration initializer and data store.
If you want to run this application standalone you don't need to config zookeeper part.

Configuration
-------------
```toml
#The port that application present rest api on
rest_api_port=":8080"

#Zookeeper connection configuration for datastore and notification center usage
####this configuration is optional###
[zookeeper]
cluster_nodes="192.168.95.171:2181"
namespace="/watch" #Znode which data will be saved under

#Mysql connection configuration which contain ws_channel_member table
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
check_interval=2

[ejabberd]
cluster_nodes="192.168.95.180:5222"
default_cluster="A"

#Every cluster can ONLY has Client or Component connection
#Please attention to use only one of them
[client]
username="4"
password="4"
domain="soroush.ir"
ping_interval=2
rate_limit=123123 #Send message/sec

[component]
name="announcer"
secret="announcer"
ping_interval=2
rate_limit=123123 #Send message/sec
```

Rest APIs
---------
|URI|type|Description|
|---|----|-----------|
|/v1/announce|POST|send channel message to that channel online members|

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
