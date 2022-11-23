# mongosync
mongosync is a lightweight and very fast utility that copies differential data between two mongodb instance.

It does **not** delete any data if it is deleted in the source database.

## Feature Matrix
| Feature | State | Comment |
|--|--|--|
| Better performance than v1 | :heavy_check_mark: | Single-col copies w/ M-T are **20x** as fast |
| DB Creation | :heavy_check_mark: |
| Collection Creation | :heavy_check_mark: |
| Differential Document Creation | :heavy_check_mark:|
| Single DB Scope | :heavy_check_mark: |
| Single Collection Scope | :heavy_check_mark: |
| Batch Uploads | :heavy_check_mark: | Every Insert or Replace requests is batched |
| Custom IDs (including objects) | :heavy_check_mark: |
| Multi-threaded processing | :heavy_check_mark: | Way better than before |
| Multiple workers per collections | :heavy_check_mark: | Has to be enabled on a case-by case basis |
| Configuration in file | :heavy_check_mark: | Obligatory |
| No-check if exist | :heavy_check_mark: | Config argument: no_find |
| Indexes | :heavy_check_mark: | Fully supported |
| Deleted Items Removal | :x: | Soon.. |
| Change Feeds | :x: | MongoDB only supports this on replica sets |
| Replica Sets | :x: | Not built in, probably easy to implement |
| Views | :x: | Not supported and will make copy fail, disable in config |

## Installation 

mongosync requires **go 1.19**, you can download go here: [Downloads - The Go Programming Language](https://go.dev/dl/)

Then, do this:

    #> go install github.com/sherweb/mongosync@v2.0.0-alpha1

Test with this:

    #> mongosync
    mongosync is an utility to sync two different mongodb instances

    Usage:
      mongosync [flags]
      mongosync [command]
    
    Available Commands:
      completion        Generate the autocompletion script for the specified shell
      copy              copy data from one mongodb instance to another
      generate-config   Generates a config with a given source and destination
      help              Help about any command
    
    Flags:
      -h, --help   help for mongosync

## Usage

First, you **must** run the following

    $> mongosync generate-config --source mongodb://user:pass@url:pro --destination mongodb://user:pass@url:port

This will output a config.yml, you can see a detailed sample with explanations on what all switches do here: [sample-config.yml](sample-config.yml)

To copy the configuration you just modified, run this:

    $> mongosync copy --config config.yml


Index copy doesn't copy the default index (the `_id_` one) and does not diff-copies indexes if they have been modified in either the source or destination.

## Contributing

To contribute, feel free to open PRs and/or issues
