package address

import (
	"fmt"
	"testing"
)

func TestNewEcPrivateKey(t *testing.T) {
	ecPrivate, err := NewEcPrivateKey()
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(ecPrivate)
}

func TestEcPrivateToPublic(t *testing.T) {
	ecPublic, err := EcPrivateToPublic("aede1bd68e1adcbbb6fe82909950cd09e55c3ce399f16c8ee2dd203c6bc6dd96")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if ecPublic != "0395da7361eab0b30062bf856a34b6e9ef0c9d0f8468718f77db6a3e0674d3efe6" {
		t.Fatalf("failed")
	}
}

func TestEcPublicToAddress(t *testing.T) {
	address, err := EcPublicToAddress("0395da7361eab0b30062bf856a34b6e9ef0c9d0f8468718f77db6a3e0674d3efe6", "testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if address != "TmQ333fbKv9yNa9Rq9BfEWEnmDdBtpieeXZ" {
		t.Fatalf("failed")
	}
}

func TestNewHdPrivate(t *testing.T) {
	priv, err := NewHdPrivate("testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(priv)
}

func TestNewHdDerive(t *testing.T) {
	priv0, err := NewHdDerive("tprvZUo1ZuEfLLFWfB2Mfycj6zPLW3FZUUqm6nmPhKbC22poNR2evRBATr7ViZD9Hr61S9q8eXdVGDFEVGPDctSJsqegw9tqVKbsAGB4GA8PPqG",
		0, "testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if priv0 != "tprvZXSyXpLZwVGVWUjEhhnZcS64zJ58kTCcnaF2GhmqpCLu4zQ2xTELMU19RndxdsLvbSghehBNTBeQ7diSJkxNMV64N9MKQt248dfUeG33pb1" {
		t.Fatalf("failed")
	}

	priv1, err := NewHdDerive("tprvZUo1ZuEfLLFWfB2Mfycj6zPLW3FZUUqm6nmPhKbC22poNR2evRBATr7ViZD9Hr61S9q8eXdVGDFEVGPDctSJsqegw9tqVKbsAGB4GA8PPqG",
		1, "testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if priv1 != "tprvZXSyXpLZwVGVZE8KcA7aQnMB4pruPaZR1KtfYeTQdivnZof2Lzs5Xa1mfnoqsB516SKPDRqDDEyJDS6TjoZXaoJRbDcKdHDqUMDuyEKUTSM" {
		t.Fatalf("failed")
	}

	priv2, err := NewHdDerive("tprvZUo1ZuEfLLFWfB2Mfycj6zPLW3FZUUqm6nmPhKbC22poNR2evRBATr7ViZD9Hr61S9q8eXdVGDFEVGPDctSJsqegw9tqVKbsAGB4GA8PPqG",
		2, "testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if priv2 != "tprvZXSyXpLZwVGVZvdUysRC76g6PpqDGYZPFWNQdAY16yQ8yoEmB5quLEN8KKpF71gAxKB7bnnZYv7kTeP3tQCEUHNC8aYuJjhcqKXuRWJLia1" {
		t.Fatalf("failed")
	}
}

func TestHdToEc(t *testing.T) {
	ecPriv0, err := HdToEc("tprvZXSyXpLZwVGVWUjEhhnZcS64zJ58kTCcnaF2GhmqpCLu4zQ2xTELMU19RndxdsLvbSghehBNTBeQ7diSJkxNMV64N9MKQt248dfUeG33pb1",
		"testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if ecPriv0 != "6e404c587ba6c65c41dba0310107c1f986574b664d34afb11420fcdb6ebb3aac" {
		t.Fatalf("failed")
	}

	ecPriv1, err := HdToEc("tprvZXSyXpLZwVGVZE8KcA7aQnMB4pruPaZR1KtfYeTQdivnZof2Lzs5Xa1mfnoqsB516SKPDRqDDEyJDS6TjoZXaoJRbDcKdHDqUMDuyEKUTSM",
		"testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if ecPriv1 != "db15cbb3c0fd199661ec815e630bee9ed77296ba41d61bc7557ab3faf247fc3f" {
		t.Fatalf("failed")
	}

	ecPriv2, err := HdToEc("tprvZXSyXpLZwVGVZvdUysRC76g6PpqDGYZPFWNQdAY16yQ8yoEmB5quLEN8KKpF71gAxKB7bnnZYv7kTeP3tQCEUHNC8aYuJjhcqKXuRWJLia1",
		"testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if ecPriv2 != "07f1ff30b96ce04e1014e5bef019026a75efa979c10f60634c1d3d4d89b04aa5" {
		t.Fatalf("failed")
	}
}

func TestEcPrivateToPublic2(t *testing.T) {
	ecPublic0, err := EcPrivateToPublic("6e404c587ba6c65c41dba0310107c1f986574b664d34afb11420fcdb6ebb3aac")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if ecPublic0 != "029246fab1354818c514307b6282b0ddf14727f6a503fa8e6f2dc584d60fb5f300" {
		t.Fatalf("failed")
	}

	ecPublic1, err := EcPrivateToPublic("db15cbb3c0fd199661ec815e630bee9ed77296ba41d61bc7557ab3faf247fc3f")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if ecPublic1 != "024d50e32f45dea07ffb3ddcfab8954875f433c28eaaac3a12019f1d3f38518e1f" {
		t.Fatalf("failed")
	}

	ecPublic2, err := EcPrivateToPublic("07f1ff30b96ce04e1014e5bef019026a75efa979c10f60634c1d3d4d89b04aa5")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if ecPublic2 != "0309f145c532ef82ad2911da967da5603268f35940b68f75d6eed184b00ce25e26" {
		t.Fatalf("failed")
	}
}

func TestEcPublicToAddress2(t *testing.T) {
	address0, err := EcPublicToAddress("029246fab1354818c514307b6282b0ddf14727f6a503fa8e6f2dc584d60fb5f300", "testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if address0 != "TmS1gTgg4izZgT18oKvXCXCmgggpxE3bHhZ" {
		t.Fatalf("failed")
	}

	address1, err := EcPublicToAddress("024d50e32f45dea07ffb3ddcfab8954875f433c28eaaac3a12019f1d3f38518e1f", "testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if address1 != "TmQzcaUv1Yj4F7i6shVA6VqzCsLpHKq8EkJ" {
		t.Fatalf("failed")
	}

	address2, err := EcPublicToAddress("0309f145c532ef82ad2911da967da5603268f35940b68f75d6eed184b00ce25e26", "testnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if address2 != "TmiwAt7ePLFnjLvenofiYPPXMyHfrNGDr9R" {
		t.Fatalf("failed")
	}
}
