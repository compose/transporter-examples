# Twitter

Reads the Twitter public* stream taking each message from the API and sending the resulting
JSON document on to MongoDB

twitter -s MONGOURI -dest-ns NAMESPACE -o TRUE/FALSE

-d        MongoDB URI to write to - defaults to localhost
-dest-ns  The name space to cat - defaults to ""
-v        If present, dump all documents to stdout

IMPORTANT:

A file named twitter.conf needs to be created to hold authentication for Twitter clients.

* Consumer-Key (API Key)
* Consumer Secret (API Secret)
* Access Token
* Access Token Secret

One per line, that order, no labels. Details on obtaining these tokens are on the [Twitter developer site](https://dev.twitter.com/oauth/overview).
