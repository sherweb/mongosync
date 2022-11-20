# mongosync
mongosync is a lightweight and very fast utility that copies differential data between two mongodb instance.

It does **not** delete any data if it is deleted in the source database.

## Feature Matrix
| Feature | State | Comment |
|--|--|--|
| DB Creation | :heavy_check_mark: |
| Collection Creation | :heavy_check_mark: |
| Differential Document Creation | :heavy_check_mark:|
| Single DB Scope | :heavy_check_mark: |
| Single Collection Scope | :heavy_check_mark: |
| Batch Uploads | :heavy_check_mark: | Every Insert or Replace requests is batched |
| Custom IDs (including objects) | :heavy_check_mark: |
| Multi-threaded processing | :heavy_check_mark: |
| Indexes | :heavy_check_mark: | Fully supported |
| Deleted Items Removal | :x: |
| Change Feeds | :x: | MongoDB only supports this on replica sets |
| Replica Sets | :x: | Not built in, probably easy to implement |

## Installation 

mongosync requires **go 1.19**, you can download go here: [Downloads - The Go Programming Language](https://go.dev/dl/)

Then, do this:

    #> go install github.com/sherweb/mongosync@v1.2.0

Test with this:

    #> mongosync
    mongosync is an utility to sync two different mongodb instances

    Usage:
      mongosync [flags]
      mongosync [command]
    
    Available Commands:
      completion  Generate the autocompletion script for the specified shell
      copy        copy data from one mongodb instance to another
      help        Help about any command
    
    Flags:
      -h, --help   help for mongosync

## Usage

To copy the entirety of one mongodb instance into another, run this:

    $> mongosync copy --source mongodb://user:pass@url:port --destination mongodb://user:pass@url:port

To copy a specific database, but all collections

    $> mongosync copy --source mongodb://user:pass@url:port --destination mongodb://user:pass@url:port --database DATABASE_NAME

To copy a specific collection inside a database

    $> mongosync copy --source mongodb://user:pass@url:port --destination mongodb://user:pass@url:port --database DATABASE_NAME --collection COLLECTION_NAME

You can add copy of indexes by adding the `-i` switch to any command. Example:

    $> mongosync copy -i --source mongodb://user:pass@url:port --destination mongodb://user:pass@url:port --database DATABASE_NAME --collection COLLECTION_NAME

Index copy doesn't copy the index ID and does not diff-copies indexes if they have been modified in either the source or destination

## Contributing

To contribute, feel free to open PRs and/or issues
