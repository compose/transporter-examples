# Cat-Mongo

Runs a cat operations on a MongoDB collection.

## Command

```
cat-mongo -s MONGOURI -ns NAMESPACE -o TRUE/FALSE
```

## Flags

* -s   MongoDB URI to connect to - defaults to localhost
* -ns  The name space to cat - defaults to ""
* -o   If present, tail the oplog for changes - defaults to false
