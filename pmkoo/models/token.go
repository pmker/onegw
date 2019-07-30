package models

import "strings"

type ITokenDao interface {
	GetAllTokens() []*Token
	InsertToken(*Token) error
	FindTokenBySymbol(string) *Token
}

type Token struct {
	TokenId   string `json:"tokenId"   db:"tokenId" gorm:"primary_key"`
	TokenSymbol     string `json:"tokenSymbol"     db:"tokenSymbol"`
	Type int    `json:"type" db:"type"`

}

func (Token) TableName() string {
	return "tokens"
}

var TokenDao ITokenDao
var TokenDaoPG ITokenDao

func init() {
	TokenDao = &tokenDaoPG{}
	TokenDaoPG = TokenDao
}

type tokenDaoPG struct {
}

func (tokenDaoPG) GetAllTokens() []*Token {
	var tokens []*Token
	DB.Find(&tokens)
	return tokens
}

func (tokenDaoPG) InsertToken(token *Token) error {
	return DB.Create(token).Error
}

func (tokenDaoPG) FindTokenBySymbol(symbol string) *Token {
	var token Token

	DB.Where("tokenSymbol = ?", symbol).Find(&token)
	if token.TokenSymbol == "" {
		return nil
	}

	return &token
}

func GetBaseTokenSymbol(marketID string) string {
	splits := strings.Split(marketID, "-")

	if len(splits) != 2 {
		return ""
	} else {
		return splits[0]
	}
}

