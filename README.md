# exchange-lib
The Qitmeer API/SDK for PMEER exchanges 

#### How to use


#### 1. How to use

##### Method 1:Get available uxto and manage utxo by yourself

- Get the latest blockorder through synchronizer.GetHistoryOrder

- Create sync.Options to set rpc information

- Use sync.NewSynchronizer to create a synchronizer

- Use synchronizer.Start to start the synchronization thread, the parameter is to start synchronization from a blockodrer

- Get the transaction from the return channel of synchronizer.Start, and then get the uxto from the transaction

- Get the latest blockorder through synchronizer.GetHistoryOrder

   ```
   opt := &sync.Options{
		RpcAddr: "127.0.0.1:1234",
		RpcUser: "admin",
		RpcPwd:  "123",
		Https:   false,
		TxChLen: 100,
	}
	synchronizer := sync.NewSynchronizer(opt)
	txChan, err := synchronizer.Start(&sync.HistoryOrder{0, 0})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	go func() {
		for {
			txs := <-txChan
			for _, tx := range txs {
				// save tx or uxto
                utxos := uxto.GetUxtos(&tx)
                // update utxo  has been spent
                spentTxs := utxo.GetSpentTxs(&tx) 
			}
		}
	}()
	go func() {
		for {
			historyId := synchronizer.GetHistoryOrder()
			if historyId.LastTxBlockOrder != 0 {
				// save historyID as the start id of the next synchronization
			}
			// update historyid every 10s
			time.Sleep(time.Second * 10)
		}
	}()
   ```
##### Method 2:Start the service and get the available utxo through api

```bash
    cd exchange
    ./build.sh
    cd bin
    cd linux
    ./exchange
```
- Configuration

```
    [api]
    listen="0.0.0.0:11360"
    
    [rpc]
    host="127.0.0.1:1234"
    tls=false
    admin="admin"
    password="123"
    
    [sync]
    # start sync from block order 
    start=0
    # synchronized address list
    address=[
        "TmiguDxFD7JvDRUvbcY7SK85NT7eVK4m9wL",
        "Tmax8njWbgrz4oagPBtQTzRMmDkUUPcVV3e",
        "TmhAddRJNQ4uAarPaxEmurdkxHqtFEsrSPh"
    ]
```
- Command

    |command|description|
    |:---|:----------------------|
    |  -c |  show version|
    |  -v |  clear all data records|

- Api

    |description|url|method|params|
    |:--------------- |:------------------------ |:----- |:-------|
    |get utxo |api/v1/utxo|GET|address|
    |send transaction  |api/v1/transaction |POST |raw;spent|
    |add address  |api/v1/address |POST |address|

- >Example 

  ##### api/v1/utxo
  ```json
    {
    "code": 0,
    "msg": "ok",
    "rs": {
        "balance": 899985000000000,
        "utxo": [
            {
                "txid": "759c0e3b69989736f39a5cf5ea057145e373af2f2117cc8cb06a7de0f6df0bcc",
                "vout": 1,
                "address": "Tmax8njWbgrz4oagPBtQTzRMmDkUUPcVV3e",
                "amount": 299995000000000,
                "spent": ""
            },
            {
                "txid": "93517537cfcb9b53b56ddefae24f109f49943cc85e38d9b9bc196aad94013baf",
                "vout": 1,
                "address": "Tmax8njWbgrz4oagPBtQTzRMmDkUUPcVV3e",
                "amount": 299995000000000,
                "spent": ""
            },
            {
                "txid": "9c89feabd1a85b497681cd4e6cea83abd758ff28427c5ec853a5a97e96c5f236",
                "vout": 0,
                "address": "Tmax8njWbgrz4oagPBtQTzRMmDkUUPcVV3e",
                "amount": 299995000000000,
                "spent": ""
            }
        ]
      }
   }
  ```

##### api/v1/transaction

- form
```json
{
    "raw":"01000000018cfef93a1e2564f2d970f562bbd7fbecbd393fa34495d95d435a08482d288ed500000000ffffffff0264737199800600001976a9146e88dc51b45362c2138de38a0ea506daf7e5ac7988ac00c817a8040000001976a914a3aa57548c99b54473126f2d2ef526f7031f7ec888ac00000000000000005a8acd5f016b483045022100fdbce12c4ee4f4214525d3f7e3380a6b5f1c2eb663d3273454d303316889185902205ff4817b107579c847787c8ce46164e1b5572fdef2937bec7270f2ed6afc980f012102786d472b1cb150134900be47c98a6ef9f666bc33dbfcf2619ee299b163670cb7",
    "spent":"[{\"txid\": \"759c0e3b69989736f39a5cf5ea057145e373af2f2117cc8cb06a7de0f6df0bcc\",\"vout\": 1,\"address\": \"Tmax8njWbgrz4oagPBtQTzRMmDkUUPcVV3e\",\"amount\": 299995000000000,\"spent\": \"\"},{\"txid\": \"93517537cfcb9b53b56ddefae24f109f49943cc85e38d9b9bc196aad94013baf\",\"vout\": 1,\"address\":\"Tmax8njWbgrz4oagPBtQTzRMmDkUUPcVV3e\",\"amount\": 299995000000000,\"spent\": \"\"}]"
}
```

##### api/v1/address

- form
```json
{
   "address": "TmUHh6bAdLbto9AYhodEwGZi9WY77CoBFXr"
}
```


#### 2. Sign transaction

```
        inputs := make(map[string]uint32, 0)
	outputs := make(map[string]uint64, 0)
	inputs["fa069bd82eda6b98e9ea40a575de1dc4c053d94a9901a956e13d30f6ab81413e"] = 0
	outputs["TmUQjNKPA3dLBB6ZfcKd4YSDThQ9Cqzmk5S"] = 100000000
	outputs["TmWRM7fk8SzBWvuUQv2cJ4T7nWPnNmzrbxi"] = 200000000
	txCode, err := sign.TxEncode(1, 0, nil, inputs, outputs)
	if err != nil {
		fmt.Println(err)
	} else {
		rawTx, ok := sign.TxSign(txCode, "b0985973cb08f7e0f013301a9686fe978cf1d887a8290184d39176c1a5157424", "testnet")
		if ok {
			client := rpc.NewClient(&rpc.RpcConfig{
				User:    "admin",
				Pwd:     "123",
				Address: "127.0.0.1:1234",
                Https:   false,
			})
			client.SendTransaction(rawTx)
		}
	}
```

#### 3. Address generation

##### One address per account

		ecPrivate, err := address.NewEcPrivateKey()
		if err != nil {
			fmt.Println(err)
			return
		}
		ecPublic, err := address.EcPrivateToPublic(ecPrivate)
		if err != nil {
			fmt.Println(err)
			return
		}
		address, err := address.EcPublicToAddress(ecPublic, "testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
##### Manage multiple addresses on one account

###### Generate HD private key
		priv, err := address.NewHdPrivate("testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
###### Generate child private key
		priv0, err := address.NewHdDerive(priv, 0, "testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
		priv1, err := address.NewHdDerive(priv, 1, "testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
######  Child private key to secp256k1 private key
		ecPriv0, err := address.HdToEc(priv0, "testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
		ecPriv1, err := address.HdToEc(priv1, "testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
######  secp256k1 sub-private key to public key
		ecPublic0, err := address.EcPrivateToPublic(ecPriv0)
		if err != nil {
			fmt.Println(err)
			return
		}
		ecPublic1, err := address.EcPrivateToPublic(ecPriv1)
		if err != nil {
			fmt.Println(err)
			return
		}
###### secp256k1 public key to address
		address0, err := address.EcPublicToAddress(ecPublic0, "testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
		address1, err := address.EcPublicToAddress(ecPublic1, "testnet")
		if err != nil {
			fmt.Println(err)
			return
		}
In addition to generating multiple addresses through the HD private key, multiple addresses can also be generated through the HD public key.
