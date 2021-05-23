package store

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Store interface {
	Put(context.Context, *BValue) error
	GetLast(ctx context.Context, userID int) (int, error)
	StoreUser
}

type StoreUser interface {
	AddUser(ctx context.Context, user *TUser) error
	//GetUser(ctx context.Context, id int) (*TUsers, error)
}

type BValue struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	UserID    int       `json:"user_id,omitempty"`
	Value     int       `json:"value,omitempty"`
}

type TUser struct {
	ID           int    `json:"id" bson:"id"`
	FirstName    string `json:"first_name" bson:"first_name"`
	LastName     string `json:"last_name" bson:"last_name"`         // optional
	UserName     string `json:"username" bson:"username"`           // optional
	LanguageCode string `json:"language_code" bson:"language_code"` // optional
	IsBot        bool   `json:"is_bot" bson:"is_bot"`               // optional
}

func (v *BValue) ParseValue(value string) (err error) {

	if v == nil {
		return fmt.Errorf("source object is nil, make before using")
	}

	if strings.Contains(value, ",") && strings.Count(value, ",") == 1 {
		value = strings.Replace(value, ",", ".", 1)
	}

	if strings.Contains(value, ".") {
		f32, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to convert entered value to float type, \"%s\", %s", value, err)
		}
		log.Printf("[DEBUG] converted float value: %f", f32)
		i32 := int(f32 * 1000)
		v.Value = i32
	} else {
		v64, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			return fmt.Errorf("failed to convert entered value to int type, \"%s\", %s", value, err)
		}

		v.Value = int(v64)
	}

	return
}
