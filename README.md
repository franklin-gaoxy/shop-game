# shop game

> A small game that simulates buying and selling, generating different prices every day to simulate changes and earn more by earning the price difference.

## start

**The first run requires executing: ./main.exe init --config document/config.yaml**

This will create the corresponding table and create its required content.

Please note: The database name must be: trading_game

And it needs to be created before `init`.

```shell
# compile
go build main.go
# run
./main --config document/config.yaml
```

or

```shell
./main.exe --config document/config.yaml
```



## more

> Need to replace the address of the database in the configuration file.
>
> If you want more products, please add them in the document/goods.yaml file.

> v1.1 Plan to add expiration time settings for different products.
