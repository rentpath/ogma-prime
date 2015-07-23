
## Installation

Install golang (v1.4.0 or higher) and mongodb (v3.0 or higher):

```bash
brew install go mongo
```

Start mongod in a separate terminal:

```bash
mongod --config /usr/local/etc/mongod.conf
```

or alternatively set mongod to start up at boot time:

```bash
ln -sfv /usr/local/opt/mongodb/*.plist ~/Library/LaunchAgents
launchctl load ~/Library/LaunchAgents/homebrew.mxcl.mongodb.plist
```

Install this package:

```bash
go install github.com/rentpath/ogma-prime
```

Run it:

```bash
cd /path/to/empty/dir/for/ogma
cp $GOPATH/src/github.com/rentpath/ogma-prime/data/config.json .

 # modify config.json if necessary

ogma-prime init
ogma-prime serve
```

## Cayley Bootstrap

Initialize cayley:

```bash
cayley init --db mongo --dbpath localhost:27017
cayley load --db mongo --dbpath localhost:27017 --format cquad --quads sample.nq
```

Sample queries against `sample.nq`:

```
g.V().Has("name", "California").In("is_in").All()
g.V().Has("name", "California").Out("is_in").All()
g.V("/locations/countries/usa").In("is_in").In("is_in").In("lives_in").All()
g.V("/locations/countries/usa").In("is_in").In("is_in").In("lives_in").Tag("id").Out(g.V("name")).Tag("name").All()
```

Drop the collection and reset everything:

```bash
mongo localhost:27017/cayley reset-mongo.js
```
