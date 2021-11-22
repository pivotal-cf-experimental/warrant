package warrant_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/golang-jwt/jwt"
	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SigningKey struct {
	ID           string
	Algorithm    string
	PrivateKey   *rsa.PrivateKey
	PublicKey    string
	SharedSecret string
}

var _ = Describe("Token", func() {
	var (
		keyA, keyB, keyC SigningKey
		service          warrant.TokensService
	)

	BeforeEach(func() {
		service = warrant.NewTokensService(warrant.Config{
			Host:          fakeUAA.URL(),
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})

		By("generating a signing key A", func() {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			Expect(err).NotTo(HaveOccurred())
			Expect(privateKey.Validate()).To(Succeed())

			publicASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
			Expect(err).NotTo(HaveOccurred())

			publicKey := string(pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: publicASN1,
			}))

			keyA = SigningKey{
				ID:         "some-key-id-a",
				Algorithm:  "RS256",
				PrivateKey: privateKey,
				PublicKey:  publicKey,
			}
		})

		By("generating a signing key B", func() {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			Expect(err).NotTo(HaveOccurred())
			Expect(privateKey.Validate()).To(Succeed())

			publicASN1 := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)

			publicKey := string(pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: publicASN1,
			}))

			keyB = SigningKey{
				ID:         "some-key-id-b",
				Algorithm:  "RS256",
				PrivateKey: privateKey,
				PublicKey:  publicKey,
			}
		})

		By("generating a signing key C", func() {
			keyC = SigningKey{
				ID:           "some-key-id-c",
				Algorithm:    "HS256",
				SharedSecret: "some-secret",
			}
		})
	})

	Describe("Verify", func() {
		Context("when the signing key uses RSA", func() {
			Context("when the public key is PKIX format", func() {
				It("verifies the token", func() {
					unsignedToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
						"client_id": "some-client-id",
						"user_id":   "some-user-id",
						"scope":     []string{"some-scope"},
						"iss":       "some-issuer",
					})

					unsignedToken.Header["kid"] = keyA.ID

					signedToken, err := unsignedToken.SignedString(keyA.PrivateKey)
					Expect(err).NotTo(HaveOccurred())

					token, err := service.Decode(signedToken)
					Expect(err).NotTo(HaveOccurred())
					Expect(token.KeyID).To(Equal("some-key-id-a"))

					err = token.Verify([]warrant.SigningKey{
						{
							KeyId:     keyA.ID,
							Algorithm: keyA.Algorithm,
							Value:     keyA.PublicKey,
						},
						{
							KeyId:     keyB.ID,
							Algorithm: keyB.Algorithm,
							Value:     keyB.PublicKey,
						},
						{
							KeyId:     keyC.ID,
							Algorithm: keyC.Algorithm,
							Value:     keyC.SharedSecret,
						},
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the public key is PKCS1 format", func() {
				It("verifies the token", func() {
					unsignedToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
						"client_id": "some-client-id",
						"user_id":   "some-user-id",
						"scope":     []string{"some-scope"},
						"iss":       "some-issuer",
					})

					unsignedToken.Header["kid"] = keyB.ID

					signedToken, err := unsignedToken.SignedString(keyB.PrivateKey)
					Expect(err).NotTo(HaveOccurred())

					token, err := service.Decode(signedToken)
					Expect(err).NotTo(HaveOccurred())
					Expect(token.KeyID).To(Equal("some-key-id-b"))

					err = token.Verify([]warrant.SigningKey{
						{
							KeyId:     keyA.ID,
							Algorithm: keyA.Algorithm,
							Value:     keyA.PublicKey,
						},
						{
							KeyId:     keyB.ID,
							Algorithm: keyB.Algorithm,
							Value:     keyB.PublicKey,
						},
						{
							KeyId:     keyC.ID,
							Algorithm: keyC.Algorithm,
							Value:     keyC.SharedSecret,
						},
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		It("verifies tokens to signed using HMAC", func() {
			unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"client_id": "some-client-id",
				"user_id":   "some-user-id",
				"scope":     []string{"some-scope"},
				"iss":       "some-issuer",
			})

			unsignedToken.Header["kid"] = keyC.ID

			signedToken, err := unsignedToken.SignedString([]byte(keyC.SharedSecret))
			Expect(err).NotTo(HaveOccurred())

			token, err := service.Decode(signedToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(token.KeyID).To(Equal("some-key-id-c"))

			err = token.Verify([]warrant.SigningKey{
				{
					KeyId:     keyA.ID,
					Algorithm: keyA.Algorithm,
					Value:     keyA.PublicKey,
				},
				{
					KeyId:     keyB.ID,
					Algorithm: keyB.Algorithm,
					Value:     keyB.PublicKey,
				},
				{
					KeyId:     keyC.ID,
					Algorithm: keyC.Algorithm,
					Value:     keyC.SharedSecret,
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			Context("when the token was not signed by a known signing key", func() {
				It("returns an error", func() {
					unsignedToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
						"client_id": "some-client-id",
						"user_id":   "some-user-id",
						"scope":     []string{"some-scope"},
						"iss":       "some-issuer",
					})

					unsignedToken.Header["kid"] = keyA.ID

					signedToken, err := unsignedToken.SignedString(keyA.PrivateKey)
					Expect(err).NotTo(HaveOccurred())

					token, err := service.Decode(signedToken)
					Expect(err).NotTo(HaveOccurred())

					err = token.Verify([]warrant.SigningKey{
						{
							KeyId:     keyB.ID,
							Algorithm: keyB.Algorithm,
							Value:     keyB.PublicKey,
						},
					})
					Expect(err).To(MatchError("token was not signed by a known key"))
				})
			})

			Context("when the token was tampered with", func() {
				It("returns an error", func() {
					unsignedToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
						"client_id": "some-client-id",
						"user_id":   "some-user-id",
						"scope":     []string{"some-scope"},
						"iss":       "some-issuer",
					})

					unsignedToken.Header["kid"] = keyA.ID

					signedToken, err := unsignedToken.SignedString(keyA.PrivateKey)
					Expect(err).NotTo(HaveOccurred())

					By("tampering with the signature by removing the base64 character", func() {
						signedToken = signedToken[:len(signedToken)-2]
					})

					token, err := service.Decode(signedToken)
					Expect(err).NotTo(HaveOccurred())

					err = token.Verify([]warrant.SigningKey{
						{
							KeyId:     keyA.ID,
							Algorithm: keyA.Algorithm,
							Value:     keyA.PublicKey,
						},
					})
					Expect(err).To(MatchError("crypto/rsa: verification error"))
				})
			})

			Context("when the RSA public key cannot be parsed", func() {
				It("returns an error", func() {
					unsignedToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
						"client_id": "some-client-id",
						"user_id":   "some-user-id",
						"scope":     []string{"some-scope"},
						"iss":       "some-issuer",
					})

					unsignedToken.Header["kid"] = keyA.ID

					signedToken, err := unsignedToken.SignedString(keyA.PrivateKey)
					Expect(err).NotTo(HaveOccurred())

					token, err := service.Decode(signedToken)
					Expect(err).NotTo(HaveOccurred())

					err = token.Verify([]warrant.SigningKey{
						{
							KeyId:     keyA.ID,
							Algorithm: keyA.Algorithm,
							Value:     "garbage public key",
						},
					})
					Expect(err).To(MatchError("public key is not valid PEM encoding"))
				})
			})

			Context("when the token algorithm is not supported", func() {
				It("returns an error", func() {
					unsignedToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
						"client_id": "some-client-id",
						"user_id":   "some-user-id",
						"scope":     []string{"some-scope"},
						"iss":       "some-issuer",
					})

					unsignedToken.Header["kid"] = keyA.ID

					signedToken, err := unsignedToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
					Expect(err).NotTo(HaveOccurred())

					token, err := service.Decode(signedToken)
					Expect(err).NotTo(HaveOccurred())

					err = token.Verify([]warrant.SigningKey{
						{
							KeyId:     keyA.ID,
							Algorithm: keyA.Algorithm,
							Value:     keyA.PublicKey,
						},
					})
					Expect(err).To(MatchError("unsupported token signing method: none"))
				})
			})
		})
	})
})
