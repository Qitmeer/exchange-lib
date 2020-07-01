package address

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/encode/base58"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/crypto/bip32"
	"github.com/Qitmeer/qitmeer/qx"
)

const (
	bip32_ByteSize = 78 + 4
)

func NewEcPrivateKey() (string, error) {
	entropyStr, err := qx.NewEntropy(32 * 8)
	if err != nil {
		return "", fmt.Errorf("new entropy error : %s", err.Error())
	}
	return qx.EcNew("secp256k1", entropyStr)
}

func EcPrivateToPublic(ecPrivate string) (string, error) {
	key, err := qx.EcPrivateKeyToEcPublicKey(false, ecPrivate)
	if err != nil {
		return "", err
	}
	return key, nil
}

func EcPublicToAddress(ecPublic string, network string) (string, error) {
	data, err := hex.DecodeString(ecPublic)
	if err != nil {
		return "", err
	}
	h := hash.Hash160(data)

	ver := qx.QitmeerBase58checkVersionFlag{}
	if err := ver.Set(network); err != nil {
		return "", err
	}
	address := base58.QitmeerCheckEncode(h, ver.Ver)
	return address, nil
}

func NewHdPrivate(network string) (string, error) {
	entropyStr, err := qx.NewEntropy(32 * 8)
	if err != nil {
		return "", fmt.Errorf("new entropy error : %s", err.Error())
	}
	entropy, err := hex.DecodeString(entropyStr)
	if err != nil {
		return "", err
	}
	//Bip32VersionFlag
	ver := qx.Bip32VersionFlag{}
	if err = ver.Set(network); err != nil {
		return "", err
	}
	masterKey, err := bip32.NewMasterKey2(entropy, ver.Version)
	if err != nil {
		return "", err
	}
	return masterKey.String(), nil
}

func HdPrivateToPublic(private string, network string) (string, error) {
	ver := qx.Bip32VersionFlag{}
	if err := ver.Set(network); err != nil {
		return "", err
	}
	data := base58.Decode(private)
	masterKey, err := bip32.Deserialize2(data, ver.Version)
	if err != nil {
		return "", err
	}
	if !masterKey.IsPrivate {
		return "", fmt.Errorf("%s is not a HD (BIP32) private key", private)
	}
	pubKey := masterKey.PublicKey()
	return pubKey.String(), nil
}

func NewHdDerive(hdPrivateOrPublic string, index uint32, network string) (string, error) {
	data := base58.Decode(hdPrivateOrPublic)
	if len(data) != bip32_ByteSize {
		return "", fmt.Errorf("invalid bip32 key size (%d), the size hould be %d", len(data), bip32_ByteSize)
	}
	ver := qx.Bip32VersionFlag{}
	if err := ver.Set(network); err != nil {
		return "", err
	}
	mKey, err := bip32.Deserialize2(data, ver.Version)
	if err != nil {
		return "", nil
	}

	childKey, err := mKey.NewChildKey(index)
	if err != nil {
		return "", err
	}
	return childKey.String(), nil
}

func HdToEc(hdPrivateOrPublic string, network string) (string, error) {
	ver := qx.Bip32VersionFlag{}
	if err := ver.Set(network); err != nil {
		return "", err
	}
	data := base58.Decode(hdPrivateOrPublic)
	key, err := bip32.Deserialize2(data, ver.Version)
	if err != nil {
		return "", err
	}
	if key.IsPrivate {
		return fmt.Sprintf("%x", key.Key[:]), nil
	} else {
		return fmt.Sprintf("%x", key.PublicKey().Key[:]), nil
	}
}
