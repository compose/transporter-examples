# Seed

Copies a namespace from one MongoDB to another, optionally tailing the oplog to keep the two in sync.

seed -s MONGOURI -d MONGOURI -source-ns NAMESPACE -dest-ns NAMESPACE -o TRUE/FALSE -v true/false

-s            Source MongoDB URI to read from - defaults to localhost
-d            Destination MongoDB URI to write to -defaults to localhost
-source-ns    The source namespace to copy - defaults to ""
-dest-ns      The source namespace to copy - defaults to ""
-o            If present, tail the oplog for changes
-v            If present, dumps all documents to stdout
