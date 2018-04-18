
Announcer
=========
Announcer is an application that send xml message to client with client or component connection ([XEP-0114](http://xmpp.org/extensions/xep-0114.html)) that present some REST API to configure and send messages.
This app initialize from toml file and zookeeper and will store it's runtime configuration in zookeeper for distribution purpose

Configuration
-------------
```toml
#Rest api port that application listen on
rest_api_port=":8080"

#Zookeeper service configuration for datastor and notification center usage
[zookeeper]
cluster_nodes="192.168.95.171:2181"
#Znode which data with be store under
namespace="/watch"

#Mysql server which contain ws_channel_member table
[mysql]
address="127.0.0.1:3306"
username="root"
password="123"
db="test"

#Redis service configuration to get online users from
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


#Only client or component can be init in init run so please attention to use only one of them
[client]
username="4"
password="4"
domain="soroush.ir"
ping_interval=2
rate_limit=123123 #sent message/sec


[component]
name="announcer"
secret="announcer"
ping_interval=2
rate_limit=123123 #sent message/sec
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