package securitykey

import (
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/go-piv/piv-go/piv"
	"github.com/smlx/piv-agent/internal/pinentry"
)

// A SecurityKey is a physical hardware token which implements PIV, such as a
// Yubikey. It provides a convenient abstraction around the low-level
// piv.YubiKey object.
type SecurityKey struct {
	card           string
	serial         uint32
	yubikey        *piv.YubiKey
	signingKeys    []SigningKey
	decryptingKeys []DecryptingKey
	cryptoKeys     []CryptoKey
}

// CryptoKey represents a cryptographic key on a hardware security device.
type CryptoKey struct {
	SlotSpec SlotSpec
	Public   crypto.PublicKey
}

// New returns a security key identified by card string.
func New(card string) (*SecurityKey, error) {
	yk, err := piv.Open(card)
	if err != nil {
		return nil, fmt.Errorf(`couldn't open card "%s": %v`, card, err)
	}
	serial, err := yk.Serial()
	if err != nil {
		return nil, fmt.Errorf(`couldn't get serial for card "%s": %v`,
			card, err)
	}

	sks, err := signingKeys(yk)
	if err != nil {
		return nil, fmt.Errorf(`couldn't get signing keys for card "%s": %v`,
			card, err)
	}
	var cryptoKeys []CryptoKey
	for _, k := range sks {
		cryptoKeys = append(cryptoKeys, k.CryptoKey)
	}

	dks, err := decryptingKeys(yk)
	if err != nil {
		return nil, fmt.Errorf(`couldn't get decrypting keys for card "%s": %v`,
			card, err)
	}
	for _, k := range dks {
		cryptoKeys = append(cryptoKeys, k.CryptoKey)
	}
	return &SecurityKey{
		card:           card,
		serial:         serial,
		yubikey:        yk,
		signingKeys:    sks,
		decryptingKeys: dks,
		cryptoKeys:     cryptoKeys,
	}, nil
}

// Retries returns the number of attempts remaining to enter the correct PIN.
func (k *SecurityKey) Retries() (int, error) {
	return k.yubikey.Retries()
}

// Serial returns the serial number of the SecurityKey.
func (k *SecurityKey) Serial() uint32 {
	return k.serial
}

// SigningKeys returns the slice of cryptographic signing keys held by the
// SecurityKey.
func (k *SecurityKey) SigningKeys() []SigningKey {
	return k.signingKeys
}

// DecryptingKeys returns the slice of cryptographic decrypting keys held by
// the SecurityKey.
func (k *SecurityKey) DecryptingKeys() []DecryptingKey {
	return k.decryptingKeys
}

// CryptoKeys returns the slice of cryptographic signing and decrypting keys
// held by the SecurityKey.
func (k *SecurityKey) CryptoKeys() []CryptoKey {
	return k.cryptoKeys
}

// PrivateKey returns the private key of the given public signing key.
func (k *SecurityKey) PrivateKey(c *CryptoKey) (crypto.PrivateKey, error) {
	return k.yubikey.PrivateKey(c.SlotSpec.Slot, c.Public,
		piv.KeyAuth{PINPrompt: pinentry.GetPin(k)})
}

// Close closes the underlying yubikey.
func (k *SecurityKey) Close() error {
	return k.yubikey.Close()
}

// AttestationCertificate returns the attestation certificate of the underlying
// yubikey.
func (k *SecurityKey) AttestationCertificate() (*x509.Certificate, error) {
	return k.yubikey.AttestationCertificate()
}

// Card returns the card identifier.
func (k *SecurityKey) Card() string {
	return k.card
}
