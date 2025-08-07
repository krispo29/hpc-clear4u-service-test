package auth

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

// const accessTokenDuration = time.Duration(time.Minute * 5)
const accessTokenDuration = time.Duration(time.Hour * 24 * 7)
const refreshTokenDuration = time.Duration(time.Hour * 24 * 7)

func init() {

	// key, err := ioutil.ReadFile("key.rsa")
	key, err := ioutil.ReadFile("private.pem")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		log.Fatal(err)
	}

	// key, err = ioutil.ReadFile("key.pub")
	key, err = ioutil.ReadFile("public.pem")
	if err != nil {
		log.Fatal(err)
	}
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(key)
	if err != nil {
		log.Fatal(err)
	}

}

type tokenClaim struct {
	UUID       string `json:"uuid"`
	IsAdmin    bool   `json:"isAdmin"`
	Authorized bool   `json:"authorized"`
	jwt.StandardClaims
}

func CreateToken(signedData *GetSignInModel, expiresIn time.Duration) (string, error) {
	// claims := jwt.MapClaims{}
	// claims["authorized"] = true
	// claims["user_id"] = user_id
	// claims["exp"] = time.Now().Add(time.Hour * 1).Unix() //Token expires after 1 hour
	// // token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	expiresAt := int64(0) // not expires
	now := time.Now()
	if expiresIn > 0 {
		expiresAt = now.Add(expiresIn).Unix()
	}

	var isAdmin bool
	// if !StringInSlice("admin", roles) {
	// 	isAdmin = false
	// } else {
	// 	isAdmin = true
	// }

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaim{
		signedData.UUID,
		isAdmin,
		true,
		jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: expiresAt,
		},
	})
	return token.SignedString(privateKey)

}

func TokenValid(r *http.Request) (*tokenClaim, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// SigningMethodHMAC
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*tokenClaim); ok && token.Valid {
		Pretty(claims)
		return claims, nil
	}
	return nil, nil
}

func ExtractToken(r *http.Request) string {
	keys := r.URL.Query()
	token := keys.Get("token")
	if token != "" {
		return token
	}
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ValidateToken(token string) (*tokenClaim, error) {
	tok, err := jwt.ParseWithClaims(token, &tokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Check is token use correct signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// return secret for this signing method
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tok.Claims.(*tokenClaim); ok && tok.Valid {
		return claims, nil
	}
	return nil, errors.New("Invalid Token")
}

func ExtractTokenID(r *http.Request) (int64, error) {

	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["user_id"]), 10, 32)
		if err != nil {
			return 0, err
		}
		return int64(uid), nil
	}
	return 0, nil
}

func Pretty(data interface{}) {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(b))
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
