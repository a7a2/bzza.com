package services

import (
	//	"fmt"
	"reflect"
	"time"

	"bzza.com/models"

	"github.com/google/uuid"
)

func RenewInstance(m interface{}) interface{} {
	return map[string]bool{"success": true}
}

func CreateInstance(m interface{}) interface{} {
	session := models.Engine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return &models.Result{Code: 500, Message: err.Error()}
	}

	var uuidDb uuid.UUID
	t := reflect.TypeOf(m)
	switch t.String() {
	case "*models.CreateInstance": //购买api服务
		has, err := models.Engine.Table("aliyun_accounts").Where("ali_uid = ? ", m.(*models.CreateInstance).AliUid).Limit(1).Cols("uuid").Get(&uuidDb)
		if err != nil {
			return map[string]string{"instanceId": m.(*models.CreateInstance).OrderBizId}
		}

		var account models.AliyunAccounts
		if has == false { //账号不存在即插入账号信息
			account = models.AliyunAccounts{
				Uuid:    uuid.Must(uuid.NewRandom()),
				Mobile:  m.(*models.CreateInstance).Mobile,
				AliUid:  m.(*models.CreateInstance).AliUid,
				Created: time.Now(),
			}
			_, err = models.Engine.Omit("uuid").Insert(&account)
			if err != nil {
				return &models.Result{Code: 500, Message: err.Error()}
			}
			uuidDb = account.Uuid
		}

		//fmt.Println(uuidDb[:])
		//插入购买api服务信息
		inApiServices := models.ApiServices{
			Account:    uuidDb.String(),
			OrderBizId: m.(*models.CreateInstance).OrderBizId,
			OrderId:    m.(*models.CreateInstance).OrderId,
			SkuId:      m.(*models.CreateInstance).SkuId,
			HasReq:     0,
			Bill:       2,
			Created:    time.Now(),
			ExpiredOn:  time.Now().AddDate(1, 0, 0),
			IsDelete:   false,
		}
		//检查api服务是否存在
		has, err = models.Engine.Table("api_services").Where("account = ? and is_delete = false and bill = ? ", uuidDb.String(), inApiServices.Bill).Limit(1).Exist()
		if err != nil {
			return &models.Result{Code: 500, Message: err.Error()}
		}
		if has == false { //不存在即插入购买的服务信息
			_, err = models.Engine.Insert(&inApiServices)
			if err != nil {
				return &models.Result{Code: 500, Message: err.Error()}
			}
		}
	}
	//fmt.Println(t.String())

	session.Commit()

	return map[string]string{"instanceId": m.(*models.CreateInstance).OrderBizId}
}
