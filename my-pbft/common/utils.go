package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func GenerateDigest(msg interface{}) []byte{
	bmsg, _ := json.Marshal(msg)
	hash := sha256.Sum256(bmsg)
	return hash[:]
}

func SignMessage(msg interface{}, privkey crypto.PrivKey ) ([]byte, error){
	dig := GenerateDigest(msg)
	sig, err := privkey.Sign(dig)
	if err != nil {
		return nil, err
	}
	return sig, nil
}


func VerifyDigest(msg interface{}, digest string) bool{
	return hex.EncodeToString(GenerateDigest(msg)) == digest
}

func VerifySignatrue(msg interface{}, sig []byte, pubkey crypto.PubKey) (bool, error){
	dig := GenerateDigest(msg)
	result, err := pubkey.Verify(dig, sig)
	if err != nil {
		return false, err
	}
	return result, nil
}