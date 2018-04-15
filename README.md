
Announcer
=========
Announcer is an application that send xml message to client with client or component connection ([XEP-0114](http://xmpp.org/extensions/xep-0114.html)) that present some REST API to configure and send messages.
This app initialize from toml file and zookeeper and will store it's runtime configuration in zookeeper for distribution purpose


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