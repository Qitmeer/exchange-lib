# exchange-lib
The Qitmeer API/SDK for PMEER exchanges 

#### How to use


#### 1. How to synchronize utxo

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
